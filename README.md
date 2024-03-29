# 项目名称：ShopCommentHub



## 需求描述
在电商评论场景中，客户畅所欲言，可发表、查看、回复评价，尽情分享购物心得。商家轻松发布商品，查看客户评价，并回应互动。审核有权删除不当评价，维护良好交流环境。

### 核心域
1. 消费者和商家评论微服务
2. 审核微服务


### 支撑域

1. 微服务内的事件总线
2. 微服务之间的MQ
3. 微服务内的存储



### 领域命令

客户可以：发表评价，查看评价，回复评价

商家可以：发布商品，查看评价，回复评价

审核可以：删除评价

#### 领域命令和领域事件

![](https://raw.githubusercontent.com/lenny-mo/PictureUploadFolder/main/领域事件图.png)


#### 领域事件的流程
消费者流程
1. 对商品SKUID=10发表评价
2. 查看商品SKUID=10的所有评价
3. 对其中的一条评价发表回复

商家流程
1. 发布商品SKUID=10
2. 获取商品SKUID=10的所有评价
3. 对其中一条评价发表回复

审核微服务: audit 
1. 对商品SKUID=10删除某条评价

## 架构设计

### 整体架构图

![image.png](https://raw.githubusercontent.com/lenny-mo/PictureUploadFolder/main/20240128115731.png)

### 请求的分层结构

![image.png](https://raw.githubusercontent.com/lenny-mo/PictureUploadFolder/main/20240128115758.png)

### 基础设施

1. 1个docker内的数据库，两个微服务共用一个数据库使用
2. 2个mongoDB, 分别给两个微服务使用，持久化本地的event
3. 1个redis，两个微服务共同使用
4. 1个kafka，系统共用
5. 1个es