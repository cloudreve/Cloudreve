如何贡献我的源代码
===

此文档介绍了 ThinkPHP 团队的组成以及运转机制，您提交的代码将给 ThinkPHP 项目带来什么好处，以及如何才能加入我们的行列。

## 通过 Github 贡献代码

ThinkPHP 目前使用 Git 来控制程序版本，如果你想为 ThinkPHP 贡献源代码，请先大致了解 Git 的使用方法。我们目前把项目托管在 GitHub 上，任何 GitHub 用户都可以向我们贡献代码。

参与的方式很简单，`fork`一份 ThinkPHP 的代码到你的仓库中，修改后提交，并向我们发起`pull request`申请，我们会及时对代码进行审查并处理你的申请。审查通过后，你的代码将被`merge`进我们的仓库中，这样你就会自动出现在贡献者名单里了，非常方便。

我们希望你贡献的代码符合：

* ThinkPHP 的编码规范
* 适当的注释，能让其他人读懂
* 遵循 Apache2 开源协议

**如果想要了解更多细节或有任何疑问，请继续阅读下面的内容**

### 注意事项

* 本项目代码格式化标准选用 [**PSR-2**](http://www.kancloud.cn/thinkphp/php-fig-psr/3141)；
* 类名和类文件名遵循 [**PSR-4**](http://www.kancloud.cn/thinkphp/php-fig-psr/3144)；
* 对于 Issues 的处理，请使用诸如 `fix #xxx(Issue ID)` 的 commit title 直接关闭 issue。
* 系统会自动在 PHP 5.4 5.5 5.6 7.0 和 HHVM 上测试修改，其中 HHVM 下的测试容许报错，请确保你的修改符合 PHP 5.4 ~ 5.6 和 PHP 7.0 的语法规范；
* 管理员不会合并造成 CI faild 的修改，若出现 CI faild 请检查自己的源代码或修改相应的[单元测试文件](tests)；

## GitHub Issue

GitHub 提供了 Issue 功能，该功能可以用于：

* 提出 bug
* 提出功能改进
* 反馈使用体验

该功能不应该用于：

 * 提出修改意见（涉及代码署名和修订追溯问题）
 * 不友善的言论

## 快速修改

**GitHub 提供了快速编辑文件的功能**

1. 登录 GitHub 帐号；
2. 浏览项目文件，找到要进行修改的文件；
3. 点击右上角铅笔图标进行修改；
4. 填写 `Commit changes` 相关内容（Title 必填）；
5. 提交修改，等待 CI 验证和管理员合并。

**若您需要一次提交大量修改，请继续阅读下面的内容**

## 完整流程

1. `fork`本项目；
2. 克隆(`clone`)你 `fork` 的项目到本地；
3. 新建分支(`branch`)并检出(`checkout`)新分支；
4. 添加本项目到你的本地 git 仓库作为上游(`upstream`)；
5. 进行修改，若你的修改包含方法或函数的增减，请记得修改[单元测试文件](tests)；
6. 变基（衍合 `rebase`）你的分支到上游 master 分支；
7. `push` 你的本地仓库到 GitHub；
8. 提交 `pull request`；
9. 等待 CI 验证（若不通过则重复 5~7，不需要重新提交 `pull request`，GitHub 会自动更新你的 `pull request`）；
10. 等待管理员处理，并及时 `rebase` 你的分支到上游 master 分支（若上游 master 分支有修改）。

*若有必要，可以 `git push -f` 强行推送 rebase 后的分支到自己的 `fork`*

*绝对不可以使用 `git push -f` 强行推送修改到上游*

### 注意事项

* 若对上述流程有任何不清楚的地方，请查阅 GIT 教程，如 [这个](http://backlogtool.com/git-guide/cn/)；
* 对于代码**不同方面**的修改，请在自己 `fork` 的项目中**创建不同的分支**（原因参见`完整流程`第9条备注部分）；
* 变基及交互式变基操作参见 [Git 交互式变基](http://pakchoi.me/2015/03/17/git-interactive-rebase/)

## 推荐资源

### 开发环境

* XAMPP for Windows 5.5.x
* WampServer (for Windows)
* upupw Apache PHP5.4 ( for Windows)

或自行安装

- Apache / Nginx
- PHP 5.4 ~ 5.6
- MySQL / MariaDB

*Windows 用户推荐添加 PHP bin 目录到 PATH，方便使用 composer*

*Linux 用户自行配置环境， Mac 用户推荐使用内置 Apache 配合 Homebrew 安装 PHP 和 MariaDB*

### 编辑器

Sublime Text 3 + phpfmt 插件

phpfmt 插件参数

```json
{
	"autocomplete": true,
	"enable_auto_align": true,
	"format_on_save": true,
	"indent_with_space": true,
	"psr1_naming": false,
	"psr2": true,
	"version": 4
}
```

或其他 编辑器 / IDE 配合 PSR2 自动格式化工具

### Git GUI

* SourceTree
* GitHub Desktop

或其他 Git 图形界面客户端
