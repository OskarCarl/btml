import logging
import sys
import traceback

import torch
from torch.types import Tensor

from model.config import DEVICE
from model.training import Model


def evaluate_model(model: Model) -> tuple[list[list[float]], list[int], list[int], list[float]]:
    """Evaluate model on test data and return predictions, true labels, and losses"""
    _ = model.model.eval()
    pred_vector_list: list[list[float]] = []
    pred_list: list[int] = []
    true_list: list[int] = []
    loss_list: list[float] = []

    loss_fn = torch.nn.CrossEntropyLoss(reduction='none')
    with torch.no_grad():
        for x, y in model.test_dataloader:
            x: Tensor = x.to(DEVICE)
            y: Tensor = y.to(DEVICE)
            pred: Tensor = model.model(x)
            pred_vector_list += pred.cpu().numpy().tolist()
            pred_list += pred.argmax(1).cpu().numpy().tolist()
            true_list += y.cpu().numpy().tolist()

            loss_list += loss_fn(pred, y).cpu().numpy().tolist()
    return pred_vector_list, pred_list, true_list, loss_list

def display_results(predictions: list[list[float]], pred_labels: list[int], true_labels: list[int], losses: list[float], limit: int = 0):
    """Display prediction results with true labels and losses"""
    correct = 0
    total = len(predictions)

    # Limit the number of results to display if specified
    display_count = limit if limit and limit < total else total

    print("\nEvaluation Results:")
    print("-" * 110)
    print(f"{'Index':<6} {'Pred 0      1      2      3      4      5      6      7      8      9  ':<73} {'True':<8} {'Correct':<8} {'Loss':<10}")
    print("-" * 110)

    for i in range(display_count):
        is_correct = pred_labels[i] == true_labels[i]

        status = "✓  " if is_correct else "  ✗"
        preds = "[" + ", ".join([f"{f:5.2f}" for f in predictions[i]]) + " ]"
        print(f"{i:<6} {preds:<73} {true_labels[i]:<8} {status:<8} {losses[i]:.6f}")

    for i in range(total):
        if pred_labels[i] == true_labels[i]:
            correct += 1

    # If we limited the display, show a message
    if limit and limit < total:
        print(f"\n... showing {display_count} of {total} results")

    # Show overall metrics
    accuracy = correct / total * 100
    print("\nOverall Metrics:")
    print(f"Accuracy: {accuracy:.2f}% ({correct}/{total})")
    print(f"Average Loss: {sum(losses)/total:.6f}")

def evaluate(model: Model, limit: int):
    try:
        predictions, pred_labels, true_labels, losses = evaluate_model(model)

        # If limit is 0, show all results
        display_count = 0 if limit == 0 else limit
        display_results(predictions, pred_labels, true_labels, losses, display_count)

    except Exception as e:
        logging.error(f"Error during evaluation: {str(e)}")
        traceback.print_exc()
        sys.exit(1)
