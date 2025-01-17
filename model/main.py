#!/usr/bin/env python

import torch, argparse
from torch import nn

from communication import ModelServer
from config import DEVICE, BATCH_SIZE, EPOCHS, LEARNING_RATE
from data import create_data_loader, print_data_shape
from training import *

def _oneshot(model, data_loader):
    # Training loop
    for t in range(EPOCHS):
        print(f"Epoch {t+1}\n-------------------------------")
        train(train_dataloader, model, loss_fn, optimizer, DEVICE)
        test(test_dataloader, model, loss_fn, DEVICE)

def main():
    print(f"Using {DEVICE} device")

    parser = argparse.ArgumentParser(description="Small neural network for fMNIST")
    parser.add_argument("--train-data", type=str, required=True, help="Path to training data file (.pt)")
    parser.add_argument("--test-data", type=str, required=True, help="Path to test data file (.pt)")
    parser.add_argument("--socket-path", type=str, help="Unix socket path for communication")
    parser.add_argument("--oneshot", type=bool, default=False, help="Only train once and exit")
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

    if args.oneshot:
        _oneshot(model, train_dataloader)
        print("Done!")
        sys.exit(0)

    # Start socket server if path is provided
    if args.socket_path:
        server = ModelServer(model, args.socket_path)
        server.start()

if __name__ == "__main__":
    main()
