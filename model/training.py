import torch
from torch import nn

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

def train(dataloader, model, loss_fn, optimizer, device):
    size = len(dataloader.dataset)
    model.train()
    for batch, (X, y) in enumerate(dataloader):
        X, y = X.to(device), y.to(device)

        # Compute prediction error
        pred = model(X)
        loss = loss_fn(pred, y)

        # Backpropagation
        loss.backward()
        optimizer.step()
        optimizer.zero_grad()

        if batch % 100 == 0:
            loss, current = loss.item(), (batch + 1) * len(X)
            print(f"loss: {loss:>7f}  [{current:>5d}/{size:>5d}]")

def test(dataloader, model, loss_fn, device):
    size = len(dataloader.dataset)
    num_batches = len(dataloader)
    model.eval()
    test_loss, correct = 0, 0
    with torch.no_grad():
        for X, y in dataloader:
            X, y = X.to(device), y.to(device)
            pred = model(X)
            test_loss += loss_fn(pred, y).item()
            correct += (pred.argmax(1) == y).type(torch.float).sum().item()
    test_loss /= num_batches
    correct /= size
    print(f"Test Error: \n Accuracy: {(100*correct):>0.1f}%, Avg loss: {test_loss:>8f} \n")

def export_model_weights(model):
    """Export model weights as a state dict."""
    return model.state_dict()

def import_model_weights(model, state_dict, weight_ratio=1.0):
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
        model.load_state_dict(state_dict)
    else:
        # Get the current state dict
        current_state_dict = model.state_dict()

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
        model.load_state_dict(averaged_state_dict)
