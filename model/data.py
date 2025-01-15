from torch.utils.data import Dataset, DataLoader
import torch

class PreparedFashionMNIST(Dataset):
    def __init__(self, data_path):
        data_dict = torch.load(data_path, weights_only=False)
        self.data = data_dict['data']
        self.labels = data_dict['labels']

    def __len__(self):
        return len(self.data)

    def __getitem__(self, idx):
        return self.data[idx], self.labels[idx]

def create_data_loader(batch_size, train_path, test_path):
    training_data, test_data = PreparedFashionMNIST(train_path), PreparedFashionMNIST(test_path)
    train_dataloader = DataLoader(training_data, batch_size=batch_size)
    test_dataloader = DataLoader(test_data, batch_size=batch_size)
    return train_dataloader, test_dataloader

def print_data_shape(dataloader):
    for X, y in dataloader:
        print(f"Shape of X [N, C, H, W]: {X.shape}")
        print(f"Shape of y: {y.shape} {y.dtype}")
        break
