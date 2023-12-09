package minecraft

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	mcingv1alpha1 "github.com/kmdkuk/mcing/api/v1alpha1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type managerProcess struct {
	k8sclient client.Client
	name      types.NamespacedName
	log       logr.Logger
	cancel    func()
}

func newManagerProcess(c client.Client, name types.NamespacedName, log logr.Logger, cancel func()) *managerProcess {
	return &managerProcess{
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
	err := p.syncWhitelist(ctx)
	if err != nil {
		return err
	}
	err = p.syncOps(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (p *managerProcess) syncWhitelist(ctx context.Context) error {
	return nil
}

func (p *managerProcess) syncOps(ctx context.Context) error {
	return nil
}

func (p *managerProcess) Cancel() {

}
