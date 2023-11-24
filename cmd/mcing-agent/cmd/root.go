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
	"log"
	"net"
	"time"

	"github.com/cybozu-go/well"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"github.com/kmdkuk/mcing/pkg/proto"
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

var config struct {
	address string
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
	RunE: func(cmd *cobra.Command, args []string) error {
		zapLogger, err := zap.NewProduction(zap.AddStacktrace(zapcore.DPanicLevel))
		if err != nil {
			return err
		}
		defer zapLogger.Sync()

		agent := server.New()

		lis, err := net.Listen("tcp", config.address)
		if err != nil {
			return err
		}
		grpcLogger := zapLogger.Named("grpc")
		grpcServer := grpc.NewServer(
			grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
				MinTime: 10 * time.Second,
			}),
			grpc.UnaryInterceptor(
				grpc_middleware.ChainUnaryServer(
					grpc_ctxtags.UnaryServerInterceptor(),
					grpc_zap.UnaryServerInterceptor(grpcLogger),
				),
			),
		)
		proto.RegisterAgentServer(grpcServer, server.NewAgentService(agent))

		well.Go(func(ctx context.Context) error {
			return grpcServer.Serve(lis)
		})
		well.Go(func(ctx context.Context) error {
			<-ctx.Done()
			grpcServer.GracefulStop()
			return nil
		})

		well.Go(func(ctx context.Context) error {
			return watcher.Watch(ctx, 10*time.Second)
		})

		if err := well.Wait(); err != nil && !well.IsSignaled(err) {
			return err
		}
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
	fs.StringVar(&config.address, "address", grpcDefaultAddr, "Listening address and port for gRPC API.")
}
