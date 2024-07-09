package design

import (
	. "goa.design/goa/v3/dsl"
)

var _ = API("network", func() {
	Title("Network Performance Monitoring Service")
	Description("Service for monitoring the performance of a network connection")
	Server("network", func() {
		Host("")
	})
})
