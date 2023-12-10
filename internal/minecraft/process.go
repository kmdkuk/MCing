package minecraft

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	mcingv1alpha1 "github.com/kmdkuk/mcing/api/v1alpha1"
	"github.com/kmdkuk/mcing/pkg/agent"
	"github.com/kmdkuk/mcing/pkg/proto"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type managerProcess struct {
	agentf    agent.AgentFactory
	k8sclient client.Client
	name      types.NamespacedName
	log       logr.Logger
	cancel    func()
}

func newManagerProcess(agentf agent.AgentFactory, c client.Client, name types.NamespacedName, log logr.Logger, cancel func()) *managerProcess {
	return &managerProcess{
		agentf:    agentf,
		k8sclient: c,
		name:      name,
		log:       log,
		cancel:    cancel,
	}
}

func (p *managerProcess) Start(ctx context.Context, interval time.Duration) {
	tick := time.NewTicker(interval)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
		case <-ctx.Done():
			p.log.Info("quit")
			return
		}

		p.log.Info("start operation")
		err := p.do(ctx)
		if err != nil {
			p.log.Error(err, "failed to operation")
			continue
		}
		p.log.Info("finish operation")
	}
}

func (p *managerProcess) do(ctx context.Context) error {
	mc := &mcingv1alpha1.Minecraft{}
	if err := p.k8sclient.Get(ctx, p.name, mc); err != nil {
		return fmt.Errorf("failed to get Minecraft: %w", err)
	}
	p.log.Info("get Minecraft", ".spec.whitelist", mc.Spec.Whitelist, ".spec.ops", mc.Spec.Ops)

	err := p.syncWhitelist(ctx, mc)
	if err != nil {
		return err
	}
	err = p.syncOps(ctx, mc)
	if err != nil {
		return err
	}
	return nil
}

func (p *managerProcess) syncWhitelist(ctx context.Context, mc *mcingv1alpha1.Minecraft) error {
	in := &proto.SyncWhitelistRequest{
		Enabled: mc.Spec.Whitelist.Enabled,
		Users:   mc.Spec.Whitelist.Users,
	}
	agent, err := p.newAgent(ctx, mc)
	if err != nil {
		return err
	}
	_, err = agent.SyncWhitelist(ctx, in)
	if err != nil {
		return err
	}
	return nil
}

func (p *managerProcess) syncOps(ctx context.Context, mc *mcingv1alpha1.Minecraft) error {
	in := &proto.SyncOpsRequest{
		Users: mc.Spec.Ops.Users,
	}
	agent, err := p.newAgent(ctx, mc)
	if err != nil {
		return err
	}
	_, err = agent.SyncOps(ctx, in)
	if err != nil {
		return err
	}
	return nil
}

func (p *managerProcess) Cancel() {
	p.cancel()
}

func (p *managerProcess) newAgent(ctx context.Context, mc *mcingv1alpha1.Minecraft) (agent.AgentConn, error) {
	pod := &corev1.Pod{}
	err := p.k8sclient.Get(ctx, client.ObjectKey{Namespace: mc.Namespace, Name: mc.PodName()}, pod)
	if err != nil {
		return nil, err
	}
	if pod.Status.PodIP == "" {
		return nil, fmt.Errorf("pod %s/%s has not been assigned an IP address", pod.Namespace, pod.Name)
	}
	return p.agentf.New(ctx, pod.Status.PodIP)
}
