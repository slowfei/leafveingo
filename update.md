
Leafveingo update log
=============

####version 0.0.1.000 rc2
1. 修改了项目名称和建立目录的约定，并更新案例和文档


-----------
####version 0.0.1.000 rc1 - 2013-10-15

目前完成了基本web framework

1. 控制器
> * 控制器返回值 text、json、模板、body out []byte、重定向、转发、out file、out html string
> * 控制器参数接收 custom struct、 LVSession.HttpSession、http.Request、http.ResponseWriter、leafveingo.HttpContext、[]uint8(Request Body)、url.URL
>

1. 路由器机制
> * URL请求解析到控制器->函数； 
> * 高级路由器机制根据调用者设计的URL进行函数和参数的解析需要实现`AdeRouterController`接口

1. 模板处理
> 模板解析、模板缓存处理、模板函数、嵌套模板函数
>

1. HttpSession
> 实现高并发获取session、自动GC清理操作、session客户端简单验证、sessionid可选随即或IPUUID、多个session超时设置（防止一次访问后创建的sesion及时清理）、cookie toke机制，form token机制

1. error 错误机制处理
> 自定义leafveingo 的自己的错误封装