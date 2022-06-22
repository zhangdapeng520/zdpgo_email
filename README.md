# zdpgo_email

使用Golang操作Email

项目地址：https://github.com/zhangdapeng520/zdpgo_email

## 版本历史

- v0.1.0 2022/04/08 常用功能
- v1.0.1 2022/05/03 新增：基于key的发送和校验
- v1.0.2 2022/05/10 优化：新增邮箱如果失败，则返回错误而非抛出错误
- v1.0.3 2022/05/10 新增：支持使用嵌入文件系统发送邮件附件
- v1.0.4 2022/05/10 BUG修复：修复嵌入文件无法正常读取的BUG
- v1.0.5 2022/05/11 BUG修复：邮箱连接失败导致异常退出
- v1.0.6 2022/05/12 新增：发送邮件并校验结果
- v1.0.7 2022/05/12 优化：代码优化
- v1.0.8 2022/05/13 升级：升级random组件为v1.1.5
- v1.0.9 2022/05/18 新增：批量发送普通附件
- v1.1.0 2022/05/24 优化：移除IMAP
- v1.1.1 2022/06/06 BUG修复：解决权限校验失败的BUG
- v1.1.2 2022/06/20 升级：日志升级
- v1.1.3 2022/06/20 升级：日志升级
- v1.1.4 2022/06/21 BUG修复：无法正常发送邮件

## 使用示例

请查看examples/basic示例，该示例演示了全部用法
