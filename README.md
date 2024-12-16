# telegram-bot-DDNS
[telegram-bot-DDNS](https://github.com/reppoor/telegram-bot-ddns)

[telegram频道](https://t.me/ddns_reppoor)
# 一款Telegram动态域名解析机器人
仅支持IPV4，IPV6请绕道

仅接受cloudflare托管的域名

![描述文本](photo.jpg)
# 开发环境
GO >= 1.21.4

MYSQL > =  5.7.34

# 功能特性
1.一键解析A记录到绑定域名

2.定时监控域名连通性进行自动切换

3.检测方法为TCP三次握手

# 准备工作

#### 如果检测对象为中国的服务器且屏蔽海外IP的机器，建议准备一台IP地区为中国的VPS

VPS操作系统建议Debian/Ubuntu


# 运行方式
#### 1.安装aaPanel
```
URL=https://www.aapanel.com/script/install_7.0_en.sh && if [ -f /usr/bin/curl ];then curl -ksSO "$URL" ;else wget --no-check-certificate -O install_7.0_en.sh "$URL";fi;bash install_7.0_en.sh aapanel
```
#### 2.进入aaPanel安装docker(如果需要使用mysql数据库，您可能需要下载mysql，创建好数据库并记住好数据库的账号密码)

#### 3.需要在宿主机root创建conf.yaml文件
```
database:
  type: "sqlite" # 数据库类型，可选mysql和sqlite,如果选mysql就需要填用户名和密码等配置
  file: "./database.db" #sqlite的文件路径，请不要更改路径
  user: "" # 数据库用户名
  password: "" # 数据库密码
  host: "" # 数据库主机
  port: "3306" # 数据库端口
  name: "" # 数据库名称
  charset: "utf8mb4" # 字符集

telegram:
  id :  #telegram用户ID
  token : "" # telegram机器人Token找@BotFather创建
  apiEndpoint: "https://api.telegram.org" #telegramAPI 可以反代，如果不知道在做什么，请不要更改

cloudflare:
  email: "" #cloudflare的email
  key: "" #cloudflare的key

network:
  enable_proxy: true # 开启:true,关闭:false。开启后一定要保证代理语法正确，否则程序报错
  proxy: "http://user:pass@127.0.0.1:7890" #配置telegram代理，支持http和socks5。示例语法 socks5://127.0.0.1:7890 账号和用户名示例语法socks5://user:pass@127.0.0.1:7890

check:
  ip_check_time : 3 # 单位秒Second
  check_time: 10 #单位分钟Minute (建议超过5分钟，否则报错)
```
#### 4.在宿主主机的root目录下创建database.db文件，名字用这个即可，创建完毕后，同时赋予该文件所有权限

(如果需要用sqlite数据库，否则忽略这条即可)
```
sudo chmod 777 /root/database.db
```
#### 5.下拉docker镜像并运行容器(使用mysql数据库的命令)
```
docker run -d -v /root/conf.yaml:/app/conf.yaml reppoor/telegram-bot-ddns:latest
```
#### 6.下拉docker镜像并运行容器(使用sqlite数据库的命令)
```
docker run -d -v /root/conf.yaml:/app/conf.yaml -v /root/database.db:/app/database.db reppoor/telegram-bot-ddns:latest
```

#### 5.启动后去容器查看日记，可以看到启动失败还是成功

# 初始化机器人

#### 找@BotFather，进入自己的机器人

1.点击Edit Bot

2.点击Edit Commands

3.输入如下命令发送
```
start - 开始
id - 获取ID
init - bot初始化
info - 转发信息
insert - 插入转发记录
check - 检测连通性
```
在docker启动后首先点击该命令，否则无法使用
```
/init 进行初始化数据库，否则无法使用
```
### Stargazers over time
[![Stargazers over time](https://starchart.cc/reppoor/telegram-bot-ddns.svg?variant=adaptive)](https://starchart.cc/reppoor/telegram-bot-ddns)