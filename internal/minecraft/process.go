package minecraft

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	mcingv1alpha1 "github.com/kmdkuk/mcing/api/v1alpha1"
	"github.com/kmdkuk/mcing/pkg/agent"
	"github.com/kmdkuk/mcing/pkg/proto"
)

type managerProcess struct {
	agentf    agent.Factory
	k8sclient client.Client
	name      types.NamespacedName
	log       logr.Logger
	cancel    func()
}

func newManagerProcess(
	agentf agent.Factory,
	c client.Client,
	name types.NamespacedName,
	log logr.Logger,
	cancel func(),
) *managerProcess {
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
			if strings.Contains(err.Error(), "has no IP") {
				p.log.Info("waiting for pod IP", "error", err)
			} else {
				p.log.Error(err, "failed to operation")
			}
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

	agent, err := p.newAgent(ctx, mc)
	if err != nil {
		return err
	}

	return p.sync(ctx, mc, agent)
}

func (p *managerProcess) sync(ctx context.Context, mc *mcingv1alpha1.Minecraft, agent agent.Conn) error {
	err := p.syncWhitelist(ctx, mc, agent)
	if err != nil {
		return err
	}
	err = p.syncOps(ctx, mc, agent)
	if err != nil {
		return err
	}
	return nil
}

func (p *managerProcess) syncWhitelist(ctx context.Context, mc *mcingv1alpha1.Minecraft, agent agent.Conn) error {
	in := &proto.SyncWhitelistRequest{
		Enabled: mc.Spec.Whitelist.Enabled,
		Users:   mc.Spec.Whitelist.Users,
	}
	p.log.Info("syncWhitelist", "in", in)
	_, err := agent.SyncWhitelist(ctx, in)
	if err != nil {
		return err
	}
	return nil
}

func (p *managerProcess) syncOps(ctx context.Context, mc *mcingv1alpha1.Minecraft, agent agent.Conn) error {
	in := &proto.SyncOpsRequest{
		Users: mc.Spec.Ops.Users,
	}
	_, err := agent.SyncOps(ctx, in)
	if err != nil {
		return err
	}
	return nil
}

func (p *managerProcess) Cancel() {
	p.cancel()
}

func (p *managerProcess) newAgent(ctx context.Context, mc *mcingv1alpha1.Minecraft) (agent.Conn, error) {
	pod := &corev1.Pod{}
	err := p.k8sclient.Get(ctx, client.ObjectKey{Namespace: mc.Namespace, Name: mc.PodName()}, pod)
	if err != nil {
		return nil, err
	}
	if pod.Status.PodIP == "" {
		return nil, fmt.Errorf("pod %s/%s has no IP", pod.Namespace, pod.Name)
	}
	return p.agentf.New(ctx, pod.Status.PodIP)
}
