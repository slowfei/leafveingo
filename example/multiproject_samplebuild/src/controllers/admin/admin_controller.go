package admin

type AdminController struct {
	tag string
}

func (a *AdminController) Index() string {
	return "hello admin"
}
