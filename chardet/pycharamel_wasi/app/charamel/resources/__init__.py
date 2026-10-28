"""
Charamel WASI resource loader.

This copy reads decompressed model files instead of upstream *.gzip resources.
"""
import pathlib
import struct
from typing import Any, Dict, List, Sequence

from charamel.encoding import Encoding

RESOURCE_DIRECTORY = pathlib.Path(__file__).parent.absolute()
WEIGHT_DIRECTORY = RESOURCE_DIRECTORY / 'weights'


def _unpack(file: pathlib.Path, pattern: str) -> List[Any]:
    with open(file, 'rb') as data:
        return [values[0] for values in struct.iter_unpack(pattern, data.read())]


def load_features() -> Dict[int, int]:
    features = _unpack(RESOURCE_DIRECTORY / 'features', pattern='>H')
    return {feature: index for index, feature in enumerate(features)}


def load_biases(encodings: Sequence[Encoding]) -> Dict[Encoding, float]:
    biases = {}
    with open(RESOURCE_DIRECTORY / 'biases', 'rb') as data:
        for line in data:
            encoding, bias = line.decode().split()
            biases[encoding] = float(bias)

    return {encoding: biases[encoding] for encoding in encodings}


def load_weights(encodings: Sequence[Encoding]) -> Dict[Encoding, List[float]]:
    weights = {}
    for encoding in encodings:
        weights[encoding] = _unpack(WEIGHT_DIRECTORY / encoding.value, pattern='>e')
    return weights
