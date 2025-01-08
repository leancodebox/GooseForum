# 一即使全，全既是一 

> one is all , all is one

本文介绍了一个用于Go语言开发约定的配置文件实现包。许多知名框架（如Laravel、Symfony和Spring）和三方包都约定了一些引导流程，使得用户能够以简便的方式引用和使用它们。 Go语言通过依赖检测（相对于Java class，Go的粒度更为细化）和 init()函数的存在，使得我们可以更加简单地实现约定模型。

本包负责读取配置文件。对于想要使用约定方式进行开发的第三方包，可以遵循以下步骤：

以封装Redis实例包为例，我们在config.toml中定义配置字段，并在Redis包中获取相应的配置：

```go
package redisManager

import (
	"github.com/leancodebox/GooseForum/app/bundles/goose/perferences"
)

var (
	config = perferences.GetExclusivePreferences()
)

var std *redisClient

func init(){
	if configIsSet(config){
		std = getNewClient(config)
    }
}

....

```

如果使用者想要使用Redis包，只需简单地引入即可。在使用第三方包并设置更多依赖时，约定方式将会变得更加方便。

例如，假设你想要开发一个基于Redis的新队列并将其封装为一个第三方包，那么在使用时只需引入Redis包即可，无需关心前两步约定的设置。在这种情况下，你只需要专注于封装队列的业务逻辑即可：

```go
package queue

import (
	"xxxx/redis"
)

var (
	defaultRedis = redis.GetStdRedis()
)

var queue *queueManager

func init(){
	if defaultRedis != nil {
		queue = getNewQueueManager(defaultRedis)
    }
}

....

```

要使用这个代码，用户只需在 config.toml 中设置必要的内容。这是一种框架/引导约定，其实现依赖于Go本身的机制，不需要编写任何boot代码（例如spring-boot、laravel-boot、symfony-boot）。 它实现了框架和模块的组合，而且采用了无专门boot代码的约定实现方式。

如果你认同这种开发方式，我们可以一起探讨如何进一步改进它。
