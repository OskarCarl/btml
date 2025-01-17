import logging
from torch.utils.data import TensorDataset, DataLoader
import torch

class PreparedFashionMNIST(TensorDataset):
    dataset: TensorDataset

    def __init__(self, data_path):
        data_dict = torch.load(data_path, weights_only=False)
        self.data = data_dict['data']
        self.labels = data_dict['labels']

    def __len__(self):
        return len(self.data)

    def __getitem__(self, idx: int):
        return self.data[idx], self.labels[idx]

def create_data_loader(batch_size: int, train_path: str, test_path: str) -> tuple[DataLoader, DataLoader]:
    training_data, test_data = PreparedFashionMNIST(train_path), PreparedFashionMNIST(test_path)
    train_dataloader = DataLoader(training_data, batch_size=batch_size, shuffle=True)
    test_dataloader = DataLoader(test_data, batch_size=batch_size, shuffle=True)
    return train_dataloader, test_dataloader

def print_data_shape(dataloader: DataLoader):
    for X, y in dataloader:
        logging.info(f"Shape of X [N, C, H, W]: {X.shape}")
        logging.info(f"Shape of y: {y.shape} {y.dtype}")
        break
