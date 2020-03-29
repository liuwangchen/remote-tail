# remote-tail

## introduce
可以tail多台机器

## config
```yaml
user: xxxxx
password: xxxx
file: 
  log: xxx.log
test:
  - xx.xx.xx.xx

online:
  - xx.xx.xx

```

mkdir ~/.remote && mv example.yaml ~/.remote/


## usage 
remote-tail {project}.{env}.{file} 这个命令可以在任意路径下执行

ps: project 是配置文件的文件名，env是配置文件中节点名可以在里边配不同环境的机器集群，file是file节点中配置的文件路径
