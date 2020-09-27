
## Daisy Redis迁移工具

仅支持Redis版本 >=2.8 的迁移

需要修改前缀时, 修改配置文件为
```bash
source:
  host:
    - 192.168.64.101:63790
  auth: Artron123
  database: 0
  prefix: 'Miss_You_'

target:
  host:
    - 192.168.64.101:63790
  auth: Artron123
  database: 0
  prefix: 'Love:You:'
```

不需要修改前缀时, 修改配置文件为
```bash
source:
  host:
    - 192.168.64.101:63790
  auth: Artron123
  database: 0
  prefix: ''

target:
  host:
    - 192.168.64.101:63790
  auth: Artron123
  database: 0
  prefix: ''
```
