package blog

type BlogController struct {
	tag string
}

func (b *BlogController) Index() string {
	return "Hello world, blog.slowfei.com"
}
