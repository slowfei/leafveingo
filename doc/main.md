Leafveingo web framework
=============

#### 框架架构与概念
	Leafveingo是一个轻量级的MVC web框架，帮助快速简洁的完成一个web或api项目。
	Leafveingo是以模块组件进行开发，要求的是高效耦合度，将特殊的模块进行划分，分开进行调用与设置。

	基本操作流程：
		URL请求->路由器解析->控制器->控制器函数->控制器返回相应(ResponseWriter)操作
		不同的Http Method会调用不同的控制器函数

	目前只是一个基础框架，有很多功能都没有实现，后续会继续实现更多便捷实用的功能模块。

#### 开发项目组织结构{#开发项目组织结构}
```
  $GOPATH
	└─src
	   └─sample				// app项目目录
	      ├─SampleWeb		// app项目名称目录，需要与编译文件同一个目录，这样编译文件就会根据项目名称查找到所需要的文件
	      │  ├─template		// 存放模板文件目录
	      │  ├─webRoot 		// web工作主目录，主要存放静态文件js、css、html...等公共访问文件，webRoot目录下都可以自定义分配安排，以下只是建议的项目规划
	      │  │  ├─images    		// (可选) 一些网站所使用的公用图片目录
	      │  │  ├─js 				// (可选) javascript 文件目录
	      │  │  └─themes 			// (可选) css主题存放目录，这样做的目的主要是可以方便切换皮肤
	      │  │     ├─core 			// (可选) css核心文件目录
	      │  │     │  └─core.css 	// (可选) css核心文件
	      │  │     └─default		// (可选) 默认css皮肤主题目录
	      │  │        ├─css			// (可选) 默认皮肤主题的css存放目录
	      │  │        │  ├─images 	// (可选) 默认皮肤主题所使用的图片目录，这样style.css访问图片路径时更好操作。
	      │  │        │  └─style.css// (可选) 默认皮肤主题的css文件
	      │  │        └─js			// (可选) 默认皮肤主题所使用的javascript文件目录
	      │  │           └─init.js 	// (可选) 默认皮肤主题所需要初始化的javascript函数或布局所使用的文件
 	      │  └─config 		// 配置文件存放目录
	      ├─src				// 存放源代码的文件夹分类（src只是为了区分源代码文件和其他资源文件)
	      │  ├─controllers	// 控制器源代码文件目录
	      │  └─models 		// 模型目录
	      └─main.go 		// 项目主文件
```
	
安装与使用
---------

使用go命令

###### 1.安装核心组件

	go get github.com/slowfei/gosfcore

###### 2.由于session使用了UUID也需要进行安装

 	go get github.com/slowfei/gosfuuid

###### 3.安装Leafveingo

	go get github.com/slowfei/leafveingo


#### 简单实例

默认使用8080端口

main.go
```golang
package main

import (
	"github.com/slowfei/leafveingo"
)

type MainController struct {
}

func (m *MainController) Index() string {
	return "Hello world, Leafvingo web framework"
}

func main() {
	leafvein := leafveingo.SharedLeafvein()
	// 需要在编译目录下，建立个与AppName一样的目录和webRoot目录，默认名称LeafveingoWeb
	// 具体可以看下(开发项目组织结构)
	leafvein.SetAppName("SampleWeb")
	leafvein.AddController("/", MainController{})
	leafvein.Start()
}

```
go build 运行后打开：http://localhost:8080/ 

输出：Hello world, Leafvingo web framework


#### [进入Leafveingo基础入门](01.md)

框架功能{#框架功能}
-------------
>1. [路由器机制](02.md)
>1. [静态文件解析](03.md)
>1. [控制器](04.md)
>1. [HttpSession](05.md)
>1. [HttpContext](06.md)
>1. [模板](07.md)

##
#### 使用协议 [LICENSE](https://github.com/slowfei/leafveingo/blob/master/LICENSE)

Leafveingo All source code is licensed under the Apache License, Version 2.0 (the "License"); 

[http://www.apache.org/licenses/LICENSE-2.0](http://www.apache.org/licenses/LICENSE-2.0) 

###