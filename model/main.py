#!/usr/bin/env python

import argparse, sys
from torch.utils.data import DataLoader

from communication import ModelServer
from config import DEVICE, BATCH_SIZE, EPOCHS
from data import create_data_loader, print_data_shape
from training import Model

def _oneshot(model: Model, data_loader: DataLoader):
    # Training loop
    for t in range(EPOCHS):
        print(f"Epoch {t+1}\n-------------------------------")
        model.train()
        model.test()

def main():
    print(f"Using {DEVICE} device")

    parser = argparse.ArgumentParser(description="Small neural network for fMNIST")
    parser.add_argument("--train-data", type=str, required=True, help="Path to training data file (.pt)")
    parser.add_argument("--test-data", type=str, required=True, help="Path to test data file (.pt)")
    parser.add_argument("--socket-path", type=str, help="Unix socket path for communication")
    parser.add_argument("--oneshot", action='store_true', help="Only train once and exit")
    args = parser.parse_args()

    # Setup data
    print(f"Loading data from {args.train_data} and {args.test_data}")
    train_dataloader, test_dataloader = create_data_loader(BATCH_SIZE, args.train_data, args.test_data)
    print_data_shape(test_dataloader)

    model = Model(train_dataloader, test_dataloader)

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
