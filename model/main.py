#!/usr/bin/env python

import torch
from torch import nn

from config import DEVICE, BATCH_SIZE, EPOCHS, LEARNING_RATE
from data import create_data_loader, print_data_shape
from training import train, test, NeuralNetwork

def main():
    print(f"Using {DEVICE} device")

    # Setup data
    train_dataloader, test_dataloader = create_data_loader(BATCH_SIZE)
    print_data_shape(test_dataloader)

    # Initialize model
    model = NeuralNetwork().to(DEVICE)
    print(model)

    # Setup training
    loss_fn = nn.CrossEntropyLoss()
    optimizer = torch.optim.SGD(model.parameters(), lr=LEARNING_RATE)

    # Training loop
    for t in range(EPOCHS):
        print(f"Epoch {t+1}\n-------------------------------")
        train(train_dataloader, model, loss_fn, optimizer, DEVICE)
        test(test_dataloader, model, loss_fn, DEVICE)
    print("Done!")

if __name__ == "__main__":
    main()
