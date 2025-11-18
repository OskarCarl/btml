import logging
from typing import Any

import torch
from torch.utils.data import DataLoader, TensorDataset


class PreparedFashionMNIST(TensorDataset):
    dataset: TensorDataset | None = None

    def __init__(self, data_path: str):
        super().__init__()
        data_dict: dict[str, list[Any]] = torch.load(data_path, weights_only=False)
        self.data: list[Any] = data_dict['data']
        self.labels: list[Any] = data_dict['labels']

    def __len__(self): # pyright: ignore[reportImplicitOverride]
        return len(self.data)

    def __getitem__(self, idx: int): # pyright: ignore[reportImplicitOverride]
        return self.data[idx], self.labels[idx]


def create_data_loader(batch_size: int, train_path: str, test_path: str) -> tuple[DataLoader[tuple[Any, ...]]|None, DataLoader[tuple[Any, ...]]]:
    logging.info(f"Loading data from {train_path} and {test_path}")
    if train_path:
        training_data = PreparedFashionMNIST(train_path)
        train_dataloader = DataLoader(
            training_data, batch_size=batch_size, shuffle=True)
    else:
        train_dataloader = None
    test_data = PreparedFashionMNIST(test_path)
    test_dataloader = DataLoader(
        test_data, batch_size=batch_size, shuffle=True)
    return train_dataloader, test_dataloader


def print_data_shape(dataloader: DataLoader[tuple[Any, ...]]):
    for x, y in dataloader:
        logging.info(f"Shape of X [N, C, H, W]: {x.shape}")
        logging.info(f"Shape of y: {y.shape} {y.dtype}")
        break
