#!/usr/bin/env python

import argparse, sys, logging
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

def setup_logging(log_file: str | None):
    """Configure logging to write to both file and stdout."""
    handlers = []

    # Always add stdout handler
    handlers.append(logging.StreamHandler(sys.stdout))

    # Add file handler if specified, with mode='w' to overwrite
    if log_file:
        handlers.append(logging.FileHandler(log_file, mode='w'))

    # Configure root logger
    logging.basicConfig(
        level=logging.INFO,
        format='%(asctime)s - %(levelname)s - %(message)s',
        handlers=handlers
    )

def main():
    parser = argparse.ArgumentParser(description="Small neural network for fMNIST")
    parser.add_argument("--train-data", type=str, required=True, help="Path to training data file (.pt)")
    parser.add_argument("--test-data", type=str, required=True, help="Path to test data file (.pt)")
    parser.add_argument("--socket-path", type=str, help="Unix socket path for communication")
    parser.add_argument("--oneshot", action='store_true', help="Only train once and exit")
    parser.add_argument("--log-file", type=str, help="Path to log file (if not specified, logs to stdout only)")
    args = parser.parse_args()

    # Setup logging
    setup_logging(args.log_file)
    logging.info(f"Using {DEVICE} device")

    # Setup data
    logging.info(f"Loading data from {args.train_data} and {args.test_data}")
    train_dataloader, test_dataloader = create_data_loader(BATCH_SIZE, args.train_data, args.test_data)
    print_data_shape(test_dataloader)

    model = Model(train_dataloader, test_dataloader)

    if args.oneshot:
        _oneshot(model, train_dataloader)
        logging.info("Done!")
        sys.exit(0)

    # Start socket server if path is provided
    if args.socket_path:
        server = ModelServer(model, args.socket_path)
        server.start()

if __name__ == "__main__":
    main()
