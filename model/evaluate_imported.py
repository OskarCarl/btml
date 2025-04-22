import sys
import torch
import logging
from typing import List, Tuple

from torch.nn import CrossEntropyLoss

from training import Model
from config import DEVICE

def evaluate_model(model: Model) -> Tuple[List[List[float]], List[int], List[int], List[float], float]:
    """Evaluate model on test data and return predictions, true labels, and losses"""
    model.model.eval()
    pred_vector_list, pred_list, true_list = [], [], []
    individual_losses: torch.Tensor = torch.Tensor()

    with torch.no_grad():
        for X, y in model.test_dataloader:
            X, y = X.to(DEVICE), y.to(DEVICE)
            pred = model.model(X)
            pred_vector_list += pred.cpu().numpy().tolist()
            pred_list += pred.argmax(1).cpu().numpy().tolist()
            true_list += y.cpu().numpy().tolist()
            individual_losses += model.loss_fn(pred, y)

    avg_loss = individual_losses.mean().item()
    loss_list = individual_losses.cpu().numpy().tolist()

    return pred_vector_list, pred_list, true_list, loss_list, avg_loss

def display_results(predictions: List[List[float]], pred_labels: List[int], true_labels: List[int], losses: List[float], limit: int = 0):
    """Display prediction results with true labels and losses"""
    correct = 0
    total = len(predictions)

    # Limit the number of results to display if specified
    display_count = limit if limit and limit < total else total

    print("\nEvaluation Results:")
    print("-" * 80)
    print(f"{'Index':<6} {'Predicted':<30} {'True':<8} {'Correct':<8} {'Loss':<10}")
    print("-" * 80)

    for i in range(display_count):
        is_correct = pred_labels[i] == true_labels[i]
        if is_correct:
            correct += 1

        status = "✓" if is_correct else "✗"

        print(f"{i:<6} {predictions[i]:<30} {true_labels[i]:<8} {status:<8} {losses[i]:.6f}")

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
        # Evaluate model
        predictions, pred_labels, true_labels, losses, avg_loss = evaluate_model(model)

        # Display results
        # If limit is 0, show all results
        display_count = 0 if limit == 0 else limit
        # TODO: needs to be adapted
        display_results(predictions, pred_labels, true_labels, losses, display_count)

    except Exception as e:
        logging.error(f"Error during evaluation: {str(e)}")
        sys.exit(1)
