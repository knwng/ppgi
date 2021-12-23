from math import trunc
import os
import random
import argparse
import networkx as nx
from generators import GENERATORS
import matplotlib.pyplot as plt
from typing import Generator
import datetime

def generate_edge(node1: list, node2: list, prob: float, g: nx.Graph):
    data = [(x, y) for x in node1 for y in node2 if random.random() <= prob]
    g.add_edges_from(data)
    return data

def generate_nodes(tag: str, node_num: int, prob: float, host_g: nx.Graph, client_g: nx.Graph):
    total = list(set([GENERATORS[tag]() for _ in range(node_num)]))
    k = int(len(total) * prob)
    client_data, host_data = random.choices(total, k=k), random.choices(total, k=k)
    host_g.add_nodes_from(host_data)
    client_g.add_nodes_from(client_data)
    return client_data, host_data

def generate_graph():
    host_g = nx.Graph()
    client_g = nx.Graph()

    # add nodes
    client_ids, host_ids = generate_nodes("identity", args.node_num, args.prob, host_g, client_g)
    client_emails, host_emails = generate_nodes("email", args.node_num, args.prob, host_g, client_g)
    client_telephones, host_telephones = generate_nodes("telephone", args.node_num, args.prob, host_g, client_g)
    client_provinces, host_provinces = generate_nodes("province", args.node_num, args.prob, host_g, client_g)

    nodes = {
        'client': {
            'identity': client_ids,
            'email': client_emails,
            'telephone': client_telephones,
            'province': client_provinces
        },
        'host': {
            'identity': host_ids,
            'email': host_emails,
            'telephone': host_telephones,
            'province': host_provinces
        }
    }

    # add edges
    edges = {
        'client': {
            'identity_email': generate_edge(client_ids, client_emails, args.prob, client_g),
            'identity_telephone': generate_edge(client_ids, client_telephones, args.prob, client_g),
            'identity_province': generate_edge(client_ids, client_provinces, args.prob, client_g)
        },
        'host': {
            'identity_email': generate_edge(host_ids, host_emails, args.prob, host_g),
            'identity_telephone': generate_edge(host_ids, host_telephones, args.prob, host_g),
            'identity_province': generate_edge(host_ids, host_provinces, args.prob, host_g)
        }
    }

    return host_g, client_g, nodes, edges

def to_csv(directory: str, nodes: dict, edges: dict, st: datetime.datetime, et: datetime.datetime) -> None:
    os.makedirs(directory, exist_ok=True)

    st, et = int(st.timestamp()), int(et.timestamp())

    for tag, data in nodes.items():
        curr_time = datetime.datetime.fromtimestamp(random.randint(st, et))
        curr_time = curr_time.isoformat('T')
        with open(os.path.join(directory, f'node_{tag}.csv'), 'w') as f:
            f.writelines([f'{x},{curr_time}\n' for x in data])

    for tag, data in edges.items():
        # src_tag, dst_tag = tag.split('_')
        curr_time = datetime.datetime.fromtimestamp(random.randint(st, et))
        curr_time = curr_time.isoformat('T')
        with open(os.path.join(directory, f'edge_{tag}.csv'), 'w') as f:
            f.writelines([f"{','.join(x)},{curr_time}\n" for x in data])

def display(g: nx.Graph) -> None:
    nx.draw(g, with_labels=True, font_weight='bold')
    plt.show()

def main():
    host_g, client_g, nodes, edges = generate_graph()
    start_time = datetime.datetime.strptime(args.start_time, '%Y-%m-%dT%H:%M:%S')
    end_time = datetime.datetime.strptime(args.end_time, '%Y-%m-%dT%H:%M:%S')

    to_csv(args.host_dir, nodes['host'], edges['host'], start_time, end_time)
    to_csv(args.client_dir, nodes['client'], edges['client'], start_time, end_time)

if __name__ == '__main__':
    parser = argparse.ArgumentParser()
    # parser.add_argument('--key_node', type=str, default='', help='format: ')
    # parser.add_argument('--aux_node_list', type=str, nargs='+',
                        # help='format: name1:type1, name2:type2, ...')
    # parser.add_argument('--edge_list', type=str, nargs='+')
    parser.add_argument('-n', '--node_num', type=int, default=10)
    parser.add_argument('-p', '--prob', type=float, default=0.7)
    parser.add_argument('-hd', '--host_dir', type=str, default='./data1',
                        help='directory to save host data csv.')
    parser.add_argument('-cd', '--client_dir', type=str, default='./data2',
                        help='directory to save client data csv.')
    parser.add_argument('-st', '--start_time', type=str, required=True,
                        help='format: YYYY-MM-DDTdd:mm:ss')
    parser.add_argument('-et', '--end_time', type=str, required=True,
                        help='format: YYYY-MM-DDTdd:mm:ss')
    args = parser.parse_args()

    main()
