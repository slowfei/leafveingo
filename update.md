
Leafveingo update log
=============

####version 0.0.2.000 rc2
1. ...


####version 0.0.2.000 rc1(欠缺文档,还未进行发布)
1. 重新整理leafveingo
> * 主要整改：独立端口的Server、灵活的路由接口
> * 多项目集成，在一个端口下也可以整合多个项目
> * config.go 重新设计架构
> * template/lv_template.go 修改初始化操作
> * session/lv_session.go 修改初始化操作，并且修改sessionManager的引用；增加Free操作
> * router.go 重新设计架构
> * status.go 微修改
> * context.go 重新设计构架
> * controller.go 重新设计构架
> * parampack.go 重新设计构架
> * response_body.go 微调整
> * leafveingo.go 重新设计构架
> * (新增)lv_reflect_router.go 反射路由接口实现
> * (新增)https的支持，可与http一起使用(不同端口处理)
> * samlpe案例整改后重新修订完成
> * multiproject多项目集成案例
> * leafveingo.go 增加模版函数的自定义添加AddTemplateFunc(...)
> * leafveingo.go 增加爱设置了一个在请求前修改Request的函数. SetCorrectRequestFunc
> * leafveingo.go 针对使用代理时，Request Header的Host Key获取真实的host
> * template/lv_template.go 修改embedTempate嵌入模板函数的错误提示，template并且增加了开发模式的设置

-----------
####version 0.0.1.000 rc2
1. 路由器
> * (new) 增加URL后缀解析控制器函数操作，具体查看sample案例的samplebuild/src/controllers/router_controller.go Forum函数。

1. 模版
> * (new) 新建模版函数：string to html
> * (new) 增加导入模版时对html文件去除空格和换行符号，可以使用SetCompactHTML进行相应的设置。

1. 日志工具
> * 对日志操作进行了整改

1. 配置文件功能
> *	增加了leafveingo的配置功能

1. 控制器参数
> * 优化了控制器参数新建结构体的速度(newStructPtr)


-----------
####version 0.0.1.000 rc1 - 2013-10-19

目前完成了基本web framework

1. 控制器
> * 控制器返回值 text、json、模板、body out []byte、重定向、转发、out file、out html string
> * 控制器参数接收 custom struct、 LVSession.HttpSession、http.Request、http.ResponseWriter、leafveingo.HttpContext、[]uint8(Request Body)、url.URL
>

1. 路由器机制
> * URL请求解析到控制器->函数； 
> * 高级路由器机制根据调用者设计的URL进行函数和参数的解析需要实现`AdeRouterController`接口

1. 模板处理
> * 模板解析、模板缓存处理、模板函数、嵌套模板函数
>

1. HttpSession
> * 实现高并发获取session、自动GC清理操作、session客户端简单验证、sessionid可选随即或IPUUID、多个session超时设置（防止一次访问后创建的sesion及时清理）、cookie toke机制，form token机制

1. error 错误机制处理
> * 自定义leafveingo 的自己的错误封装




