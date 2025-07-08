//go:build world_debug_info_enabled

package world

type WorldDebugInfo struct {
	History PlayerInputArray
}

func (w *World) StepDebug(input PlayerInput) {
	w.WorldDebugInfo.History.V[w.WorldDebugInfo.History.N] = input
	w.WorldDebugInfo.History.N++
}
