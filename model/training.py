import logging
import torch
import torch
from torch import nn
from torch.utils.data import DataLoader

from config import DEVICE, LEARNING_RATE

class NeuralNetwork(nn.Module):
    def __init__(self):
        super().__init__()
        self.flatten = nn.Flatten()
        self.linear_relu_stack = nn.Sequential(
            nn.Linear(28*28, 100),
            nn.ReLU(),
            nn.Linear(100, 100),
            nn.ReLU(),
            nn.Linear(100, 10)
        )

    def forward(self, x):
        x = self.flatten(x)
        logits = self.linear_relu_stack(x)
        return logits

class Model:
    model: NeuralNetwork
    loss_fn: nn.CrossEntropyLoss
    optimizer: torch.optim.SGD
    train_dataloader: DataLoader
    test_dataloader: DataLoader

    def __init__(self, train_dataloader: DataLoader, test_dataloader: DataLoader):
        self.model = NeuralNetwork().to(DEVICE)
        logging.info(f"Initialized new model: {self.model}")

        # Setup training
        self.loss_fn = nn.CrossEntropyLoss()
        self.optimizer = torch.optim.SGD(self.model.parameters(), lr=LEARNING_RATE)

        self.train_dataloader = train_dataloader
        self.test_dataloader = test_dataloader

    def train(self):
        size = len(self.train_dataloader.dataset) #type: ignore
        self.model.train()
        for batch, (X, y) in enumerate(self.train_dataloader):
            X, y = X.to(DEVICE), y.to(DEVICE)

            # Compute prediction error
            pred = self.model(X)
            loss = self.loss_fn(pred, y)

            # Backpropagation
            loss.backward()
            self.optimizer.step()
            self.optimizer.zero_grad()

            if batch % 100 == 0:
                loss, current = loss.item(), (batch + 1) * len(X)
                logging.info(f"loss: {loss:>7f}  [{current:>5d}/{size:>5d}]")

    def test(self):
        size = len(self.test_dataloader.dataset) #type: ignore
        num_batches = len(self.test_dataloader)
        self.model.eval()
        test_loss, correct = 0, 0
        with torch.no_grad():
            for X, y in self.test_dataloader:
                X, y = X.to(DEVICE), y.to(DEVICE)
                pred = self.model(X)
                test_loss += self.loss_fn(pred, y).item()
                correct += (pred.argmax(1) == y).type(torch.float).sum().item()
        test_loss /= num_batches
        correct /= size
        logging.info(f"Test Error: \n Accuracy: {(100*correct):>0.1f}%, Avg loss: {test_loss:>8f} \n")

    def export_model_weights(self):
        """Export model weights as a state dict."""
        return self.model.state_dict()

    def import_model_weights(self, state_dict: dict, weight_ratio: float = 1.0):
        """
        Import model weights from a state dict with weighted averaging.

        Args:
            model: The model to import weights into
            state_dict: The state dict containing the weights to import
            weight_ratio: Float between 0 and 1, where:
                0 = keep current weights
                1 = use imported weights completely (default)
                values between 0-1 = weighted average of current and imported weights
        """
        if not 0 <= weight_ratio <= 1:
            raise ValueError("weight_ratio must be between 0 and 1")

        if weight_ratio == 1.0:
            # If weight_ratio is 1, just load the imported weights directly
            self.model.load_state_dict(state_dict)
        else:
            # Get the current state dict
            current_state_dict = self.model.state_dict()

            # Create a new state dict with weighted average
            averaged_state_dict = {}
            for key in current_state_dict.keys():
                if key in state_dict:
                    current_weights = current_state_dict[key]
                    imported_weights = state_dict[key]

                    # Compute weighted average
                    averaged_weights = (
                        (1 - weight_ratio) * current_weights +
                        weight_ratio * imported_weights
                    )
                    averaged_state_dict[key] = averaged_weights
                else:
                    averaged_state_dict[key] = current_state_dict[key]

            # Load the averaged weights
            self.model.load_state_dict(averaged_state_dict)
