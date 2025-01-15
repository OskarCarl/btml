#!/usr/bin/env python

import torch, argparse
from torch import nn

from config import DEVICE, BATCH_SIZE, EPOCHS, LEARNING_RATE
from data import create_data_loader, print_data_shape
from training import train, test, NeuralNetwork

def main():
    print(f"Using {DEVICE} device")

    parser = argparse.ArgumentParser(description="Small neural network for fMNIST")
    parser.add_argument("--train-data", type=str, required=True, help="Path to training data file (.pt)")
    parser.add_argument("--test-data", type=str, required=True, help="Path to test data file (.pt)")
    args = parser.parse_args()

    # Setup data
    print(f"Loading data from {args.train_data} and {args.test_data}")
    train_dataloader, test_dataloader = create_data_loader(BATCH_SIZE, args.train_data, args.test_data)
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
