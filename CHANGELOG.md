# CHANGELOG

## Dockyard v0.7.0
新特性增量开发及Bug修复
#### 新特性
* 支持organization/team/user三级用户权限管理模型 

#### Bug修复
* 修正上传两个相同tag的镜像后者的镜像信息未覆盖前者的信息问题 

## Dockyard v0.6.0
新特性增量开发
#### 新特性
* 支持ceph对象存储驱动
* 支持查询docker镜像功能
* 支持删除docker镜像功能
* 兼容docker API规范
* backend驱动框架使能
* 镜像存储去本地化
* 解除对wrench依赖

#### Bug修复
* N/A

## Dockyard v0.5.0
新特性增量开发及Bug修复，切换Go 1.5.3.
#### 新特性
* 支持ACI镜像存储
* 支持Dockayrd自研OSS对象存储服务，链接：[OSS](oss/README.md)

#### Bug修复
* 修正系统测试用例问题
* 修正docker V2协议问题