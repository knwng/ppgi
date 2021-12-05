import os
import random
import argparse
import networkx as nx
from generators import GENERATORS
import matplotlib.pyplot as plt
from typing import Generator

def generate_edge(node1: list, node2: list, prob: float) -> Generator:
    return [(x, y) for x in node1 for y in node2 if random.random() <= prob]

def generate_graph():
    g = nx.Graph()

    # add nodes
    ids = list(set([GENERATORS['id']() for _ in range(args.node_num)]))
    g.add_nodes_from(ids)
    emails = list(set([GENERATORS['email']() for _ in range(args.node_num)]))
    g.add_nodes_from(emails)
    telephones = list(set([GENERATORS['telephone']() for _ in range(args.node_num)]))
    g.add_nodes_from(telephones)
    provinces = list(set([GENERATORS['province']() for _ in range(args.node_num)]))
    g.add_nodes_from(provinces)
    nodes = {'id': ids, 'email': emails,
             'telephone': telephones, 'province': provinces}

    # add edges
    id_email_edges = generate_edge(ids, emails, args.prob)
    g.add_edges_from(id_email_edges)
    id_telephone_edges = generate_edge(ids, telephones, args.prob)
    g.add_edges_from(id_telephone_edges)
    id_province_edges = generate_edge(ids, provinces, args.prob)
    g.add_edges_from(id_province_edges)
    edges = {'id_email': id_email_edges, 'id_telephone': id_telephone_edges,
             'id_province': id_province_edges}

    return g, nodes, edges

def to_csv(nodes: dict, edges: dict) -> None:
    os.makedirs(args.dir, exist_ok=True)
    for tag, data in nodes.items():
        with open(os.path.join(args.dir, f'node_{tag}.csv'), 'w') as f:
            f.writelines([x + '\n' for x in data])
    
    for tag, data in edges.items():
        src_tag, dst_tag = tag.split('_')
        with open(os.path.join(args.dir, f'edge_{tag}.csv'), 'w') as f:
            f.writelines([','.join(x) + '\n' for x in data])

def display(g: nx.Graph) -> None:
    nx.draw(g, with_labels=True, font_weight='bold')
    plt.show()

def main():
    g, nodes, edges = generate_graph()
    # display(g)
    to_csv(nodes, edges)

if __name__ == '__main__':
    parser = argparse.ArgumentParser()
    # parser.add_argument('--key_node', type=str, default='', help='format: ')
    # parser.add_argument('--aux_node_list', type=str, nargs='+',
                        # help='format: name1:type1, name2:type2, ...')
    # parser.add_argument('--edge_list', type=str, nargs='+')
    parser.add_argument('-n', '--node_num', type=int, default=10)
    parser.add_argument('-p', '--prob', type=float, default=0.5)
    parser.add_argument('-d', '--dir', type=str, default='./data',
                        help='directory to save csv.')
    args = parser.parse_args()

    main()
