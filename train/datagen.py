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
                {"id": "0"},
                {"id": "1"},
            ],
            "links": [
                {"from": "0", "to": "1"},
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
                {"id": "0"},
                {"id": "1", "group": "X", "external": True},
                {"id": "2", "group": "X", "external": True},
                {"id": "3", "group": "X", "external": True},
            ],
        }
    ),
    Datapoint(
        """three connected boxes""",
        {
            "nodes": [
                {"id": "0"},
                {"id": "1"},
                {"id": "2"},
            ],
            "links": [
                {"from": "0", "to": "1", "direction": "LR"},
                {"from": "1", "to": "2", "direction": "TD"},
                {"from": "2", "to": "0", "direction": "RL"},
            ],
        }
    ),
    Datapoint(
        """c4 containers: golang web server authenticating users read from external mysql database""",
        {
            "nodes": [
                {"id": "0", "label": "Web Server", "technology": "Go", "description": "Authenticates users"},
                {"id": "1", "label": "Database", "technology": "MySQL", "external": True, "is_database": True},
            ],
            "links": [
                {"from": "0", "to": "1", "direction": "LR"},
            ],
        }
    ),
    Datapoint(
        """Five containers in three groups. First container is a Flink Application which performs feature engineering 
        using JSON encoded user behavioural clickstream consumed from AWS Kinesis Stream over HTTP. It publishes AVRO 
        encoded results to the kafka topic over TCP and infers the machine learning model by sending JSON data over 
        rest API. The Flink application is deployed to AWS KDA of the Business Domain account. Kafka topic is part of 
        the Streaming Platform, which sinks the data to the Datalake, AWS S3 bucket. The model is deployed to the MLPlatform.
        MLPlatform, clickstream and datalake belong to the Data Platform. All but Flink application are external.""",
        {
            "nodes": [
                {"id": "0", "label": "Flink Application", "technology": "AWS KDA",
                 "description": "Performs feature engineering", "group": "Business Domain account"},
                {"id": "1", "label": "User behavioural clickstream", "technology": "AWS Kinesis Stream",
                 "external": True, "is_queue": True, "group": "Data Platform"},
                {"id": "2", "label": "Kafka topic", "technology": "Kafka", "external": True, "is_queue": True,
                 "group": "Streaming Platform"},
                {"id": "3", "label": "Machine learning model", "technology": "MLPlatform", "external": True,
                 "group": "Data Platform"},
                {"id": "4", "label": "Datalake", "technology": "AWS S3", "external": True, "group": "Data Platform"},
            ],
            "links": [
                {"from": "0", "to": "1", "direction": "TD", "label": "consumes clickstream", "technology": "HTTP/JSON"},
                {"from": "0", "to": "2", "direction": "LR", "label": "publishes results", "technology": "TCP/AVRO"},
                {"from": "0", "to": "3", "direction": "TD", "label": "infers the machine learning model",
                 "technology": "HTTP/JSON"},
                {"from": "2", "to": "4", "direction": "TD", "label": "sinks the data", "technology": "HTTP/JSON"},
            ],
        }
    ),
    # TODO: complete the example for
    # TODO: http://www.plantuml.com/plantuml/uml/fLJ1Sjem43sNhr0vcKc33-cfDueXpHJID1ucqvD6bXTRGPOyaWnXEltthXpZs12OpjGdURNxtlFkMtyKYiig1P8xLzelOMZORfm9brT9PS5mhHmeD-Qw24l9bAiAUMrTAaKIJZzVF_ZGQha82sOT60pHALOmeS2CIymT3E8ztXJqgwvKoim-O5RRJsGuIRTWdB12PIJI1LOCH-JtWE3J8WHS2hwnpW0hA2gnLG66IbOaAVCGMMOWOqveHIPbYRdrUUYldcAogFE6dKImP0tCLSOVZ2v81n_PyUdpHqd0TfPoPYrJgM5q0tjRCWx-0wPljUGHk3QfFJ1_FwGDHuCFBThF2Ye8DdYqmjBcgmvwkYfJCc-YBU1h4OaRgHtKeD0foBUcBFqhkLDhCA0uN6xCWz4lU-8qSJcG6WYn-oDO2mEvHYjWUiGSasm3_be1TzeS2vmVig_YcwipF6c3yjfnClXpwf5CKf-5bRTI9qmIpzpweoUGpbvSWAq2izN2qAJY6t3zmfgi4HhV-EuT3QN6w0-cF-0pSy3e2eajII1dMw4hWunfBRNBKSCNajODnfgvWMUsNFZovOBW2hcrDdwgs8b7d6LqsVSdGfiCsxCpnlr6XyyM1p-gSLUNSSzRdJEU8xp0Nu4H1R4EJUe972yH-b-mpxk-h18fRDvPpGOvJ8H2BxJQ-pu_3zXseujtsegEh_y3gvbNNrjdrCBleUsirpjQegwMTvoyCyB_kMfV7TF7ttuyqUZN_MHvDwMBsUtiFWf6Vm40
    Datapoint(
        """""",
        {
            "nodes": [],
            "links": [
                {"from": "0", "to": "1", "direction": "TD", "label": "consumes clickstream", "technology": "HTTP/JSON"},
                {"from": "0", "to": "2", "direction": "DT", "label": "caches interim state", "technology": "TCP"},
                {"from": "0", "to": "3", "direction": "DT", "label": "publishes features", "technology": "HTTP/JSON"},
                {"from": "4", "to": "3", "direction": "DT", "label": "consumes features", "technology": "HTTP/JSON"},
                {"from": "4", "to": "5", "direction": "TD", "label": "infers the machine learning model",
                 "technology": "HTTP/JSON"},
                {"from": "6", "to": "3", "direction": "DT", "label": "consumes features", "technology": "HTTP/JSON"},
                {"from": "6", "to": "7", "direction": "TD", "label": "sinks the data", "technology": "HTTP/JSON"},
            ],
        }
    )
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
