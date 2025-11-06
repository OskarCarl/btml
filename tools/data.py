# pyright: basic
import argparse
import os

import torch
from torch.utils.data import random_split
from torchvision import datasets, transforms  # pyright: ignore[reportMissingImports]


def download_and_process_fashion_mnist(data_dir: str = "data"):
    """Download FashionMNIST dataset and apply basic transforms."""
    transform = transforms.Compose([
        transforms.ToTensor(),
        transforms.Normalize((0.5,), (0.5,))
    ])

    # Download training data
    train_dataset = datasets.FashionMNIST(
        root=data_dir,
        train=True,
        download=True,
        transform=transform
    )

    # Download test data
    test_dataset = datasets.FashionMNIST(
        root=data_dir,
        train=False,
        download=True,
        transform=transform
    )

    return train_dataset, test_dataset


def split_and_save_dataset(dataset, num_splits: int, output_dir: str, prefix: str):
    """Split a dataset into multiple parts and save them to disk."""
    # Calculate split sizes
    total_size = len(dataset)
    split_size = total_size // num_splits
    split_sizes = [split_size] * (num_splits - 1)
    split_sizes.append(total_size - split_size *
                       (num_splits - 1))  # Account for remainder

    # Split the dataset
    splits = random_split(dataset, split_sizes)

    # Save each split
    for idx, split in enumerate(splits):
        # Extract the actual data and labels for this split
        split_data = []
        split_labels = []
        for i in split.indices:
            data, label = dataset[i]
            split_data.append(data)
            split_labels.append(label)

        # Stack the tensors
        split_data = torch.stack(split_data)
        split_labels = torch.tensor(split_labels)

        # Create a dictionary with the split data
        split_dict = {
            'data': split_data,
            'labels': split_labels
        }

        output_path = os.path.join(output_dir, f"fMNIST_{
                                   prefix}_split_{idx}.pt")
        torch.save(split_dict, output_path)
        print(f"Saved split {idx} with {len(split)} samples to {output_path}")


def main():
    parser = argparse.ArgumentParser(
        description="Prepare FashionMNIST dataset")
    parser.add_argument("--num-splits", type=int, default=5,
                        help="Number of splits to create for each set")
    parser.add_argument("--data-dir", type=str, default="data",
                        help="Directory to store the raw dataset")
    parser.add_argument("--output-dir", type=str, default="processed_data",
                        help="Directory to store the processed splits")

    args = parser.parse_args()

    # Download and process the dataset
    train_dataset, test_dataset = download_and_process_fashion_mnist(
        args.data_dir)

    # Split and save training data
    split_and_save_dataset(
        train_dataset,
        args.num_splits,
        args.output_dir,
        "train"
    )

    # Split and save test data
    split_and_save_dataset(
        test_dataset,
        args.num_splits,
        args.output_dir,
        "test"
    )


if __name__ == "__main__":
    main()
