"""Script to generate jsonl file with the train data to fine-tune OpenAI model."""
import argparse
import json
import time
from dataclasses import dataclass
from typing import List


@dataclass
class Datapoint:
    prompt: str
    completion: dict

    def to_json(self) -> str:
        return json.dumps({
            "prompt": self.prompt,
            "completion": json.dumps(self.completion),
        })


class Sample(List[Datapoint]):
    def to_jsonl(self):
        return "\n".join([i.to_json() for i in self])


gen_input = Sample([
    Datapoint(
        "two nodes, one link",
        {
            "nodes": [
                {"id": "node0"},
                {"id": "node1"},
            ],
            "links": [
                {"from": "node0", "to": "node1"},
            ]
        },
    ),
    Datapoint(
        """diagram with three nodes producer, broker and consumer, and links: producer to broker from left to right, 
and consumer to broker from right to left. 
title: foobar, footer: current date. 
nodes: producer, broker, consumer. 
producer's technology is Go and label is Account, consumer's technology is Java and label is Moderation. 
broker is kafka.
consumer and broker are external.""",
        {
            "title": "foobar",
            "footer": "current date",
            "nodes": [
                {"id": "producer", "label": "Account", "technology": "Go", "description": "Publishes account events"},
                {"id": "broker", "label": "Kafka", "technology": "Kafka", "description": "Broker", "external": True},
                {"id": "consumer", "label": "Moderation", "technology": "Java",
                 "description": "Consumes account events", "external": True, "is_queue": True},
            ],
            "links": [
                {"from": "producer", "to": "broker", "label": "Publishes events", "direction": "LR"},
                {"from": "consumer", "to": "broker", "label": "Consumes events", "direction": "RL"},
            ]
        },
    ),
    Datapoint(
        """Draw c4 container diagram with four containers, thee of which are external and belong to the system X.""",
        {
            "nodes": [
                {"id": "node0"},
                {"id": "node1", "system": "X", "external": True},
                {"id": "node2", "system": "X", "external": True},
                {"id": "node3", "system": "X", "external": True},
            ],
        }
    ),
    Datapoint(
        """three connected boxes""",
        {
            "nodes": [
                {"id": "node0"},
                {"id": "node1"},
                {"id": "node2"},
            ],
            "links": [
                {"from": "node0", "to": "node1", "direction": "LR"},
                {"from": "node1", "to": "node2", "direction": "TD"},
                {"from": "node2", "to": "node0", "direction": "RL"},
            ],
        }
    ),
])


def args() -> argparse.Namespace:
    """Parses stdin arguments."""

    parser = argparse.ArgumentParser(description="OpenAI training sample generator.")
    parser.add_argument("-o", "--output", required=False, help="Path to store output.",
                        default=f"/tmp/openai_{time.strftime('%Y%m%dT%H%M%SZ.jsonl', time.gmtime())}")
    return parser.parse_args()


def main(path: str):
    """Program entrypoint.

    Args:
        path: Path to store output.
    """
    with open(path, "w") as fw:
        fw.write(gen_input.to_jsonl())


if __name__ == "__main__":
    main(args().output)
