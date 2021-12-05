# 构建随机关系图

## 点，边定义
- 点:
  - id: 关键节点，这里模仿身份证号，由18位数字组成
  - email: 非关键节点，由随机可打印字符+'@gmail.com'组成
  - telephone: 非关键节点，由11位数字组成，首位一定为1
  - province: 非关键节点，字符串，如Shandong, Beijing
- 边：为了简化关系，只存在id结点到其他结点之间的边，命名规则为`<src_id>_<drc_id>`
  - id_email
  - id_telephone
  - id_province

## 使用
### 生成随机图
```bash
python generate_random_graph.py -n 每类结点数量 -p 边生成概率 -d 生成图数据存放目录
```

### 部署nebula-graph和nebula-studio
```bash
./deploy_graph_and_studio.sh
```

### 下载、编译nebula-importer并向nebula-graph导入数据
```bash
./import_graph.sh
```

## TODO
[ ] 提高图生成脚本的通用性
