package download

import (
	"context"
	"errors"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/util/exec"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	mcingv1alpha1 "github.com/kmdkuk/mcing/api/v1alpha1"
	agent "github.com/kmdkuk/mcing/pkg/proto"
)

// MockKubeExecutor mocks kube.Executor.
type MockKubeExecutor struct {
	mock.Mock
}

//nolint:errcheck // mock implementation
func (m *MockKubeExecutor) PortForward(
	namespace, podName string,
	remotePort int,
	out, errOut io.Writer,
) (int, chan struct{}, error) {
	args := m.Called(namespace, podName, remotePort, out, errOut)
	return args.Int(0), args.Get(1).(chan struct{}), args.Error(2)
}

func (m *MockKubeExecutor) Exec(
	ctx context.Context,
	namespace, podName, container string,
	cmd []string,
	stdin io.Reader,
	out, errOut io.Writer,
) error {
	args := m.Called(ctx, namespace, podName, container, cmd, stdin, out, errOut)
	return args.Error(0)
}

// MockAgentClient mocks AgentClient.
type MockAgentClient struct {
	mock.Mock
	agent.AgentClient // Embed interface
}

//nolint:errcheck // mock implementation
func (m *MockAgentClient) SaveOff(
	ctx context.Context,
	in *agent.SaveOffRequest,
	opts ...grpc.CallOption,
) (*agent.SaveOffResponse, error) {
	args := m.Called(ctx, in, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*agent.SaveOffResponse), args.Error(1)
}

//nolint:errcheck // mock implementation
func (m *MockAgentClient) SaveAllFlush(
	ctx context.Context,
	in *agent.SaveAllFlushRequest,
	opts ...grpc.CallOption,
) (*agent.SaveAllFlushResponse, error) {
	args := m.Called(ctx, in, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*agent.SaveAllFlushResponse), args.Error(1)
}

//nolint:errcheck // mock implementation
func (m *MockAgentClient) SaveOn(
	ctx context.Context,
	in *agent.SaveOnRequest,
	opts ...grpc.CallOption,
) (*agent.SaveOnResponse, error) {
	args := m.Called(ctx, in, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*agent.SaveOnResponse), args.Error(1)
}

//nolint:funlen // test function
func TestDownloader_Run(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = mcingv1alpha1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	tests := []struct {
		name        string
		minecraft   *mcingv1alpha1.Minecraft
		setupMocks  func(*MockKubeExecutor, *MockAgentClient)
		expectedErr bool
	}{
		{
			name: "Success",
			minecraft: &mcingv1alpha1.Minecraft{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-mc",
					Namespace: "default",
				},
				Spec: mcingv1alpha1.MinecraftSpec{
					Backup: mcingv1alpha1.Backup{
						Excludes: []string{"logs"},
					},
				},
			},
			setupMocks: func(mk *MockKubeExecutor, _ *MockAgentClient) {
				// Check sleeping status
				mk.On("Exec", mock.Anything, "default", "mcing-test-mc-0", "minecraft",
					[]string{"pgrep", "java"},
					mock.Anything, mock.Anything, mock.Anything).
					Return(nil)

				stopCh := make(chan struct{})
				// PortForward
				mk.On("PortForward", "default", "mcing-test-mc-0", 9080, mock.Anything, mock.Anything).
					Return(12345, stopCh, nil)

				// Exec (tar)
				mk.On("Exec", mock.Anything, "default", "mcing-test-mc-0", "minecraft",
					[]string{"tar", "czf", "-", "-C", "/data", "--exclude", "session.lock", "--exclude", "logs", "."},
					mock.Anything, mock.Anything, mock.Anything).
					Return(nil)
			},
			expectedErr: false,
		},
		{
			name: "PortForward Failure",
			minecraft: &mcingv1alpha1.Minecraft{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-mc",
					Namespace: "default",
				},
				Spec: mcingv1alpha1.MinecraftSpec{
					Backup: mcingv1alpha1.Backup{
						Excludes: []string{"logs"},
					},
				},
			},
			setupMocks: func(mk *MockKubeExecutor, _ *MockAgentClient) {
				// Check sleeping status
				mk.On("Exec", mock.Anything, "default", "mcing-test-mc-0", "minecraft",
					[]string{"pgrep", "java"},
					mock.Anything, mock.Anything, mock.Anything).
					Return(nil)

				mk.On("PortForward", "default", "mcing-test-mc-0", 9080, mock.Anything, mock.Anything).
					Return(0, (chan struct{})(nil), errors.New("portforward failed"))
			},
			expectedErr: true,
		},
		{
			name: "Sleeping Server (AutoPause Enabled)",
			minecraft: &mcingv1alpha1.Minecraft{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-mc",
					Namespace: "default",
				},
				Spec: mcingv1alpha1.MinecraftSpec{
					AutoPause: mcingv1alpha1.AutoPause{
						Enabled: new(bool),
					},
					Backup: mcingv1alpha1.Backup{
						Excludes: []string{"logs"},
					},
				},
			},
			setupMocks: func(mk *MockKubeExecutor, _ *MockAgentClient) {
				// 1. checkSleepingAndWarn -> isServerSleeping -> Exec("pgrep java")
				// Returns error "exit code 1" mimicking process not found
				mk.On("Exec", mock.Anything, "default", "mcing-test-mc-0", "minecraft",
					[]string{"pgrep", "java"},
					mock.Anything, mock.Anything, mock.Anything).
					Return(&exec.CodeExitError{Err: errors.New("command terminated with exit code 1"), Code: 1})

				// 2. PortForward is SKIPPED because server is sleeping.

				// 3. Fallback to Download -> Exec("tar")
				mk.On("Exec", mock.Anything, "default", "mcing-test-mc-0", "minecraft",
					[]string{"tar", "czf", "-", "-C", "/data", "--exclude", "session.lock", "--exclude", "logs", "."},
					mock.Anything, mock.Anything, mock.Anything).
					Return(nil)
			},
			expectedErr: false,
		},
		{
			name: "Not Sleeping Server (Connection Error)",
			minecraft: &mcingv1alpha1.Minecraft{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-mc",
					Namespace: "default",
				},
				Spec: mcingv1alpha1.MinecraftSpec{
					AutoPause: mcingv1alpha1.AutoPause{
						Enabled: new(bool),
					},
					Backup: mcingv1alpha1.Backup{
						Excludes: []string{"logs"},
					},
				},
			},
			setupMocks: func(mk *MockKubeExecutor, _ *MockAgentClient) {
				// 1. PortForward fails
				mk.On("PortForward", "default", "mcing-test-mc-0", 9080, mock.Anything, mock.Anything).
					Return(0, (chan struct{})(nil), errors.New("connection refused"))

				// 2. checkSleepingAndWarn -> isServerSleeping -> Exec("pgrep java")
				// Returns nil (success), meaning process found
				mk.On("Exec", mock.Anything, "default", "mcing-test-mc-0", "minecraft",
					[]string{"pgrep", "java"},
					mock.Anything, mock.Anything, mock.Anything).
					Return(nil)
			},
			expectedErr: true,
		},
		{
			name: "SaveOff Failure",
			minecraft: &mcingv1alpha1.Minecraft{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-mc",
					Namespace: "default",
				},
				Spec: mcingv1alpha1.MinecraftSpec{
					Backup: mcingv1alpha1.Backup{
						Excludes: []string{"logs"},
					},
				},
			},
			setupMocks: func(mk *MockKubeExecutor, ma *MockAgentClient) {
				stopCh := make(chan struct{})
				mk.On("PortForward", "default", "mcing-test-mc-0", 9080, mock.Anything, mock.Anything).
					Return(12345, stopCh, nil)

				ma.On("SaveOff", mock.Anything, mock.Anything, mock.Anything).
					Return(nil, errors.New("saveoff failed"))

				// Check sleeping status
				mk.On("Exec", mock.Anything, "default", "mcing-test-mc-0", "minecraft",
					[]string{"pgrep", "java"},
					mock.Anything, mock.Anything, mock.Anything).
					Return(nil)
			},
			expectedErr: true,
		},
		{
			name: "Exec Failure (tar)",
			minecraft: &mcingv1alpha1.Minecraft{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-mc",
					Namespace: "default",
				},
				Spec: mcingv1alpha1.MinecraftSpec{
					Backup: mcingv1alpha1.Backup{
						Excludes: []string{"logs"},
					},
				},
			},
			setupMocks: func(mk *MockKubeExecutor, ma *MockAgentClient) {
				// Check sleeping status
				mk.On("Exec", mock.Anything, "default", "mcing-test-mc-0", "minecraft",
					[]string{"pgrep", "java"},
					mock.Anything, mock.Anything, mock.Anything).
					Return(nil)

				stopCh := make(chan struct{})
				mk.On("PortForward", "default", "mcing-test-mc-0", 9080, mock.Anything, mock.Anything).
					Return(12345, stopCh, nil)

				ma.On("SaveOff", mock.Anything, mock.Anything, mock.Anything).
					Return(&agent.SaveOffResponse{}, nil)
				ma.On("SaveAllFlush", mock.Anything, mock.Anything, mock.Anything).
					Return(&agent.SaveAllFlushResponse{}, nil)

				mk.On("Exec", mock.Anything, "default", "mcing-test-mc-0", "minecraft",
					mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(errors.New("tar failed"))

				// Important: SaveOn must be called even if Exec fails
				ma.On("SaveOn", mock.Anything, mock.Anything, mock.Anything).
					Return(&agent.SaveOnResponse{}, nil)
			},
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.minecraft.Spec.AutoPause.Enabled != nil {
				// Helper to set enabled to true for pointers
				val := true
				tt.minecraft.Spec.AutoPause.Enabled = &val
			}

			fakeClient := fake.NewClientBuilder().
				WithScheme(scheme).
				WithObjects(tt.minecraft).
				Build()

			mockKube := new(MockKubeExecutor)
			mockAgent := new(MockAgentClient)

			if tt.setupMocks != nil {
				tt.setupMocks(mockKube, mockAgent)
			}

			// Add expectations for default success flows if not overridden
			if tt.name == "Success" {
				mockAgent.On("SaveOff", mock.Anything, mock.Anything, mock.Anything).
					Return(&agent.SaveOffResponse{}, nil).
					Maybe()
				mockAgent.On("SaveAllFlush", mock.Anything, mock.Anything, mock.Anything).
					Return(&agent.SaveAllFlushResponse{}, nil).
					Maybe()
				mockAgent.On("SaveOn", mock.Anything, mock.Anything, mock.Anything).
					Return(&agent.SaveOnResponse{}, nil).
					Maybe()
			}

			opts := &Options{
				Namespace:     "default",
				MinecraftName: "test-mc",
				Output:        "test-output.tar.gz",
			}

			d := NewDownloader(opts, fakeClient, mockKube)
			d.agentFactory = func(_ int) (agent.AgentClient, func() error, error) {
				return mockAgent, func() error { return nil }, nil
			}

			err := d.Run(context.Background())
			if tt.expectedErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			// Always cleanup output file
			_ = os.Remove("test-output.tar.gz")
			mockKube.AssertExpectations(t)
			mockAgent.AssertExpectations(t)
		})
	}
}
