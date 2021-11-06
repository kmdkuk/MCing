package cmd

import (
	mcingv1alpha1 "github.com/kmdkuk/mcing/api/v1alpha1"
)

func getPodName(mc *mcingv1alpha1.Minecraft) string {
	return mc.PrefixedName() + "-0"
}
