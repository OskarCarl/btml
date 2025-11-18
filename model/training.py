import logging
from typing import Any

import torch
from numpy import array, concatenate, unique
from torch import nn
from torch.types import Tensor
from torch.utils.data import DataLoader

from model.config import DEVICE, LEARNING_RATE


# Based on https://github.com/Abhi-H/CNN-with-Fashion-MNIST-dataset/
class NeuralNetwork(nn.Module):
    def __init__(self):
        super().__init__() # pyright: ignore[reportUnknownMemberType]
        self.cnn1: nn.Conv2d = nn.Conv2d(in_channels=1,out_channels=16,kernel_size=5,stride=1,padding=2)
        self.relu1: nn.ELU = nn.ELU()
        _ = nn.init.xavier_uniform_(self.cnn1.weight)

        self.maxpool1: nn.MaxPool2d = nn.MaxPool2d(kernel_size=2)

        self.cnn2: nn.Conv2d = nn.Conv2d(in_channels=16,out_channels=32,kernel_size=5,stride=1,padding=2)
        self.relu2: nn.ELU = nn.ELU()
        _ = nn.init.xavier_uniform_(self.cnn2.weight)

        self.maxpool2: nn.MaxPool2d = nn.MaxPool2d(kernel_size=2)

        self.fcl: nn.Linear = nn.Linear(32*7*7,10)

    def forward(self, x: Tensor): # pyright: ignore[reportImplicitOverride]
        out: Tensor = self.cnn1(x)
        out = self.relu1(out)
        out = self.maxpool1(out)
        out = self.cnn2(out)
        out = self.relu2(out)
        out = self.maxpool2(out)

        out = out.view(out.size(0),-1)

        out = self.fcl(out)

        return out


class Model:
    model: NeuralNetwork
    loss_fn: nn.CrossEntropyLoss
    optimizer: torch.optim.SGD
    train_dataloader: DataLoader[tuple[Tensor, ...]]|None
    test_dataloader: DataLoader[tuple[Tensor, ...]]

    def __init__(self, train_dataloader: DataLoader[tuple[Any, ...]]|None, test_dataloader: DataLoader[tuple[Any, ...]]):
        self.model = NeuralNetwork().to(DEVICE)
        logging.info("Initialized new model")

        # Setup training
        self.loss_fn = nn.CrossEntropyLoss()
        self.optimizer = torch.optim.SGD(
            self.model.parameters(), lr=LEARNING_RATE)

        self.train_dataloader = train_dataloader
        self.test_dataloader = test_dataloader

    def train(self) -> float:
        """
        Trains the model for an epoch.

        Returns:
            float: The average loss over all batches
        """
        assert self.train_dataloader is not None, "train_dataloader is None"
        size = len(self.train_dataloader.dataset) # pyright: ignore[reportArgumentType]
        losses: list[float] = []
        _ = self.model.train()
        for batch, (x, y) in enumerate(self.train_dataloader):
            x, y = x.to(DEVICE), y.to(DEVICE)

            # Compute prediction error
            pred = self.model(x)
            loss: Tensor = self.loss_fn(pred, y)

            # Backpropagation
            _ = loss.backward() # pyright: ignore[reportUnknownMemberType]
            _ = self.optimizer.step() # pyright: ignore[reportUnknownMemberType, reportUnknownVariableType]
            self.optimizer.zero_grad()

            if batch % 100 == 0:
                loss_val: float = loss.item()
                current = (batch + 1) * len(x)
                logging.info(f"loss: {loss_val:>7f}  [{current:>5d}/{size:>5d}]")
                losses += [loss_val]
        return sum(losses)/len(losses)

    def test(self) -> tuple[float, float, dict[int, float]]:
        """
        Evaluates the model.

        Returns:
            float: accuracy
            float: loss
            dict[int, float]: the relative prevalence of each label in the generated predictions
        """
        size = len(self.test_dataloader.dataset) # pyright: ignore[reportArgumentType]
        num_batches = len(self.test_dataloader)
        _ = self.model.eval()
        test_loss, correct = 0, 0
        pred_labels = array([], dtype=int)
        with torch.no_grad():
            for x, y in self.test_dataloader:
                x, y = x.to(DEVICE), y.to(DEVICE)
                pred = self.model(x)
                test_loss += self.loss_fn(pred, y).item()
                correct += (pred.argmax(1) == y).type(torch.float).sum().item()
                pred_labels = concatenate((pred_labels, pred.argmax(1).numpy().astype(int)))
        pred_labels, pred_counts = unique(pred_labels, return_counts=True)
        guesses = {int(label): pred_counts[i]/size for i, label in enumerate(pred_labels) if pred_counts[i] > size/13}
        test_loss /= num_batches
        correct /= size
        logging.info(
            f"Test Error: Accuracy: {correct:>0.4f}, Avg loss: {test_loss:>8f}")
        return correct, test_loss, guesses

    def export_model_weights(self) -> dict[str, Any]:
        """Export model weights as a state dict."""
        return self.model.state_dict()

    def import_model_weights(self, state_dict: dict[str, Any], weight_ratio: float = 1.0):
        """
        Import model weights from a state dict with weighted averaging.

        Args:
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
            _ = self.model.load_state_dict(state_dict)
        else:
            # Get the current state dict
            current_state_dict = self.model.state_dict()

            # Create a new state dict with weighted average
            averaged_state_dict: dict[str, Tensor] = {}
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
            _ = self.model.load_state_dict(averaged_state_dict)
