// whatsapp-bridge-project/bridge-go/dispatch.go

package main

import (
	"whatsapp-bridge/internal/voip"
	"whatsapp-bridge/internal/wabridge"
)

// router decides which subsystem owns each raw whatsmeow event.
type router struct {
	wa      *wabridge.Service
	sidecar *voip.Sidecar
}

// newRouter wires the event router.
func newRouter(wa *wabridge.Service, sidecar *voip.Sidecar) *router {
	return &router{wa: wa, sidecar: sidecar}
}

// handle gives call events to the VoIP sidecar and everything else to the
// WhatsApp service.
func (rt *router) handle(evt any) {
	if rt.sidecar != nil && rt.sidecar.HandleEvent(evt) {
		return
	}
	rt.wa.Dispatch(evt)
}
