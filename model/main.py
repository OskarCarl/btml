#!/usr/bin/env python

import argparse
import sys
import logging
from torch import load

from communication import ModelServer
from config import DEVICE, BATCH_SIZE, EPOCHS
from data import create_data_loader, print_data_shape
from training import Model
from evaluate_imported import evaluate


def _oneshot(model: Model):
    # Training loop
    for t in range(EPOCHS):
        logging.info(f"Epoch {t+1} -------------------------------")
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
    parser = argparse.ArgumentParser(
        description="Small neural network for fMNIST")
    parser.add_argument("--train-data", type=str,
                        help="Path to training data file (.pt)")
    parser.add_argument("--test-data", type=str, required=True,
                        help="Path to test data file (.pt)")
    parser.add_argument("--socket", type=str,
                        help="Unix socket path for communication")
    parser.add_argument("--oneshot", action='store_true',
                        help="Only train once and exit")
    parser.add_argument("--log-file", type=str,
                        help="Path to log file (if not specified, logs to stdout only)")
    parser.add_argument("--weights", type=str,
                        help="Path to the saved model weights file (.pt or .pth)")
    parser.add_argument("--evaluate", action='store_true',
                        help="Evaluate the model and exit")
    parser.add_argument("--limit", type=int, default=20,
                        help="Limit the number of results to display in evaluation mode (default: 20, 0 for all)")
    args = parser.parse_args()

    if (not args.train_data and not args.evaluate):
        logging.error("You need to either provide train data or enable evaluation mode")
        sys.exit(2)

    if args.oneshot and args.socket:
        logging.error("Cannot use --oneshot and --socket together")
        sys.exit(3)

    # Setup logging
    setup_logging(args.log_file)
    logging.info(f"Using {DEVICE} device")

    # Setup data
    train_dataloader, test_dataloader = create_data_loader(
        BATCH_SIZE, args.train_data, args.test_data)
    print_data_shape(test_dataloader)

    model = Model(train_dataloader, test_dataloader)
    if args.weights:
        model.model.load_state_dict(load(args.weights, weights_only=True))
    if args.evaluate:
        evaluate(model, args.limit)
        logging.info("Done!")
    elif args.socket:
        server = ModelServer(model, args.socket)
        logging.info(f"Starting server on {args.socket}")
        server.start()
    elif args.oneshot:
        _oneshot(model)
        logging.info("Done!")
    else:
        logging.error("No action specified")
        sys.exit(4)


if __name__ == "__main__":
    main()
