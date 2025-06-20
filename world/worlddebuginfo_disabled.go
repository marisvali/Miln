//go:build world_debug_info_disabled

package world

type WorldDebugInfo struct {
}

func (w *World) StepDebug(input PlayerInput) {
}
