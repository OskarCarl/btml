import torch

DEVICE = (
    "cuda"
    if torch.cuda.is_available()
    else "mps"
    if torch.backends.mps.is_available()
    else "cpu"
)

BATCH_SIZE = 64
EPOCHS = 5
LEARNING_RATE = 0.015
