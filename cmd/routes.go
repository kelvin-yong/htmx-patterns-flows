package main

type permission string

const (
	PermissionAP01 = permission("AP01")
	PermissionAP02 = permission("AP02")
	PermissionAP03 = permission("AP03")
	PermissionAP04 = permission("AP04")
)

// Interaction types
// - main: can load on browser, it is a GET
// - api:  interaction paths for AJAX request, does not appear on browser.
// - virtual : appears on browser, does not correspond to a real endpoint on the server. Has a fallback

type route struct {
	Path          string // path
	Method        string // method (not really in use other than documenting)
	Interaction   string // see Interaction types (not really in use other than documenting)
	Permissions   []permission
	FallbackRoute *route
}

var (
	RouteDemo3      = route{Path: "demo03", Method: "GET", Permissions: []permission{PermissionAP01}, Interaction: "main"}
	RouteDemo3Step1 = route{Path: "demo03-month", Interaction: "virtual", FallbackRoute: &RouteDemo3}
	RouteDemo3Step2 = route{Path: "demo03-fav-colour", Interaction: "virtual", FallbackRoute: &RouteDemo3}
	RouteDemo3Step3 = route{Path: "demo03-thankyou", Interaction: "virtual", FallbackRoute: &RouteDemo3}
)
