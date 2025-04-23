import sys
import torch
import logging
from typing import List, Tuple

from training import Model
from config import DEVICE
import traceback

def evaluate_model(model: Model) -> Tuple[List[List[float]], List[int], List[int], List[float]]:
    """Evaluate model on test data and return predictions, true labels, and losses"""
    model.model.eval()
    pred_vector_list, pred_list, true_list, loss_list = [], [], [], []

    loss_fn = torch.nn.CrossEntropyLoss(reduction='none')
    with torch.no_grad():
        for X, y in model.test_dataloader:
            X, y = X.to(DEVICE), y.to(DEVICE)
            pred = model.model(X)
            pred_vector_list += pred.cpu().numpy().tolist()
            pred_list += pred.argmax(1).cpu().numpy().tolist()
            true_list += y.cpu().numpy().tolist()

            loss_list += loss_fn(pred, y).cpu().numpy().tolist()
    return pred_vector_list, pred_list, true_list, loss_list

def display_results(predictions: List[List[float]], pred_labels: List[int], true_labels: List[int], losses: List[float], limit: int = 0):
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
