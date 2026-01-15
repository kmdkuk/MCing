/*
Copyright © 2020 kouki kamada(kmdkuk.com)

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"context"
	"fmt"
	"log"
	"net"
	"path"
	"sync"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	james4krcon "github.com/james4k/rcon"
	"github.com/kmdkuk/mcing/pkg/config"
	"github.com/kmdkuk/mcing/pkg/constants"
	"github.com/kmdkuk/mcing/pkg/proto"
	"github.com/kmdkuk/mcing/pkg/rcon"
	"github.com/kmdkuk/mcing/pkg/server"
	"github.com/kmdkuk/mcing/pkg/watcher"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

const (
	grpcDefaultAddr = ":9080"
)

var flags struct {
	address string
}

// InterceptorLogger adapts zap logger to interceptor logger.
// This code is simple enough to be copied and not imported.
func InterceptorLogger(l *zap.Logger) logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		f := make([]zap.Field, 0, len(fields)/2)

		for i := 0; i < len(fields); i += 2 {
			key := fields[i]
			value := fields[i+1]

			switch v := value.(type) {
			case string:
				f = append(f, zap.String(key.(string), v))
			case int:
				f = append(f, zap.Int(key.(string), v))
			case bool:
				f = append(f, zap.Bool(key.(string), v))
			default:
				f = append(f, zap.Any(key.(string), v))
			}
		}

		logger := l.WithOptions(zap.AddCallerSkip(1)).With(f...)

		switch lvl {
		case logging.LevelDebug:
			logger.Debug(msg)
		case logging.LevelInfo:
			logger.Info(msg)
		case logging.LevelWarn:
			logger.Warn(msg)
		case logging.LevelError:
			logger.Error(msg)
		default:
			panic(fmt.Sprintf("unknown level %v", lvl))
		}
	})
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "mcing-agent",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		zapLogger, err := zap.NewProduction(zap.AddStacktrace(zapcore.DPanicLevel))
		if err != nil {
			return err
		}
		defer func() {
			err = zapLogger.Sync()
		}()

		lis, err := net.Listen("tcp", flags.address)
		if err != nil {
			return err
		}
		grpcLogger := zapLogger.Named("grpc")
		opts := []logging.Option{
			logging.WithLogOnEvents(logging.StartCall, logging.FinishCall),
			// Add any other option (check functions starting with logging.With).
		}
		grpcServer := grpc.NewServer(
			grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
				MinTime: 10 * time.Second,
			}),
			grpc.ChainUnaryInterceptor(
				logging.UnaryServerInterceptor(InterceptorLogger(grpcLogger), opts...),
				// Add any other interceptor you want.
			),
			grpc.ChainStreamInterceptor(
				logging.StreamServerInterceptor(InterceptorLogger(grpcLogger), opts...),
				// Add any other interceptor you want.
			),
		)
		retryCount := 0
		var conn *james4krcon.RemoteConsole
		for {
			props, err := config.ParseServerPropsFromPath(path.Join(constants.DataPath, constants.ServerPropsName))
			if err != nil {
				return err
			}

			hostPort := "localhost:" + props[constants.RconPortProps]
			password := props[constants.RconPasswordProps]

			conn, err = rcon.NewConn(hostPort, password)
			if err == nil {
				break
			}
			if retryCount > 10 {
				return err
			}
			retryCount++
			wait := 1 * retryCount
			zapLogger.Error(fmt.Sprintf("connection error, retry after %d seconds", wait), zap.Error(err))
			time.Sleep(time.Duration(wait) * time.Second)
		}
		defer func() {
			err = conn.Close()
		}()

		proto.RegisterAgentServer(grpcServer, server.NewAgentService(zapLogger, conn))

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := grpcServer.Serve(lis)
			if err != nil {
				zapLogger.Error("failed to serve", zap.Error(err))
				cancel()
			}
		}()

		wg.Add(1)
		go func(ctx context.Context) {
			defer wg.Done()
			<-ctx.Done()
			grpcServer.GracefulStop()
		}(ctx)

		wg.Add(1)
		go func(ctx context.Context) {
			defer wg.Done()
			err := watcher.Watch(ctx, conn, 10*time.Second)
			if err != nil {
				zapLogger.Error("failed to watch", zap.Error(err))
			}
		}(ctx)

		wg.Wait()
		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func init() {
	fs := rootCmd.Flags()
	fs.StringVar(&flags.address, "address", grpcDefaultAddr, "Listening address and port for gRPC API.")
}
