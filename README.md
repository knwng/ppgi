# privacy-preserving-graph-intersection

## Description
RFC: [Nebula Hackathon RFC](https://docs.google.com/document/d/1AWWKMocy6VV27nzAYF5opfXE1TdYDmN9lEcmt1uE3Cg/edit?usp=sharing)

## Usage
### 1. Deploy required middleware
- MessageQueue, pulsar is supported
- KV Database, redis is supported
- Graph Database, NebulaGraph is supported

scripts for deploying these services are located in `./tests/deploy/`

### 2. Prepare the configuration
#### Client Configuration
```yaml
role: client
conn_timeout: 60
log_file: ./client.log
algorithm:
  type: rsa
  first_hash: sha256    # algorithm-specified params
  second_hash: md5
  key_bits: 4096
graph:
  address: 192.168.31.147
  port: 9669
  username: root
  password: nebula
  graph_name: relation_graph
  graph_definition: ./conf/graph.yaml   # defined below
  neighbor_steps:
  # steps used in GO sentence to find neighbors, if only one number N is provided, will return the neighbors N steps away(the sentence will be `GO N STEPS FROM`), if two numbers N, M are provided, will return the neighbors within N to M steps(the sentence will be `GO N TO M STEPS FROM`)
    - 1
    - 2
  fetch_interval: 10
mq:
  type: pulsar
  url: pulsar://192.168.31.147:6650
  in_topic: client
  out_topic: host
  schema: ./conf/pulsar_schema.json
kv:
  type: redis
  url: localhost:6379
  password: ""
  db: 0
```

#### Host Configuration
```yaml
role: host
conn_timeout: 60
log_file: ./host.log
algorithm:
  type: rsa
  first_hash: sha256
  second_hash: md5
  key_bits: 4096
graph:
  address: 192.168.31.147
  port: 9669
  username: root
  password: nebula
  graph_name: relation_graph_host
  graph_definition: ./conf/graph.yaml
  neighbor_steps:
    - 1
    - 2
  fetch_interval: 10
mq:
  type: pulsar
  url: pulsar://192.168.31.147:6650
  in_topic: host
  out_topic: client
  schema: ./conf/pulsar_schema.json
kv:
  type: redis
  url: localhost:6379
  password: ""
  db: 0
```

#### graph structure definition
```yaml
nodes:
  - 
    type: identity
    related_edges:
      - identity_email
      - identity_telephone
      - identity_province
    props:
      - data
      - collect_time
    time_prop: collect_time
    # data_prop: data
  -
    type: email
    related_edges:
      - identity_email
    props:
      - data
      - collect_time
    time_prop: collect_time
    # data_prop: data
  -
    type: telephone
    related_edges:
      - identity_telephone
    props:
      - data
      - collect_time
    time_prop: collect_time
    # data_prop: data
  -
    type: province
    related_edges:
      - identity_province
    props:
      - data
      - collect_time
    time_prop: collect_time
    # data_prop: data
edges:
  -
    type: identity_email
    props:
      - collect_time
    time_prop: collect_time
  -
    type: identity_telephone
    props:
      - collect_time
    time_prop: collect_time
  -
    type: identity_province
    props:
      - collect_time
    time_prop: collect_time
```

Parameters:
- nodes: **Required**, list of nodes in graph db.
    - type: **Required**, the type(or tag) of node.
    - related_edges: **Required**, type of edges related to the node.
    - props: **Required**, list of node's properties.
    - time_prop: **Required**, the property that is used to filter nodes by time.
    - data_prop: **Optional**, the property that will be returned by query, if it's not set, 'VertexID' will be used.
- edges: **Required**, list of edges in graph db.
    - type: **Required**, the type of edge.
    - props: **Required**, list of edge's properties.
    - time_prop: **Required**, the property that is used to filter edges by time.

### 3. Build and run the application on both client side and host side
```bash
make build

./cmd/ppgi --config <client/host configuration file path>
```
