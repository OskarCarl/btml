import logging
import os
import socket
from concurrent.futures import ThreadPoolExecutor
from io import BytesIO
from pathlib import Path

import grpc
from torch import load, save

from model.lib.ipc import peer_model_pb2 as messages
from model.lib.ipc import peer_model_pb2_grpc as ipc
from model.training import Model


class ModelServer:
    def __init__(self, model: Model, socket_path: str):
        self.model: Model = model
        self.socket_path: str = socket_path
        self.conn: socket.socket | None = None

    def start(self):
        """Start the model server and handle connections."""
        if os.path.exists(self.socket_path):
            os.unlink(self.socket_path)

        server = grpc.server(ThreadPoolExecutor(max_workers=1))
        _ = server.add_insecure_port("unix://" + self.socket_path)

        ipc.add_TrainServicer_to_server(TrainService(self.model), server)  # pyright: ignore[reportUnknownMemberType]
        ipc.add_EvalServicer_to_server(EvalService(self.model), server)  # pyright: ignore[reportUnknownMemberType]
        ipc.add_ImportWeightsServicer_to_server(ImportWeightsService(self.model), server)  # pyright: ignore[reportUnknownMemberType]
        ipc.add_ExportWeightsServicer_to_server(ExportWeightsService(self.model), server)  # pyright: ignore[reportUnknownMemberType]

        server.start()
        logging.info("gRPC server started")

        while True:
            try:
                _ = server.wait_for_termination()
            except Exception as e:
                logging.error(f"Error when running server: {e}")
            finally:
                _ = server.stop(2)

class TrainService(ipc.TrainServicer):
    def __init__(self, model: Model) -> None:
        super().__init__()
        self.model: Model = model

    def Train(self, request: messages.TrainRequest, context) -> messages.TrainResponse:  # pyright: ignore[reportImplicitOverride]
        response = messages.TrainResponse()
        response.loss = self.model.train()
        response.success = True
        return response

class EvalService(ipc.EvalServicer):
    def __init__(self, model: Model) -> None:
        super().__init__()
        self.model: Model = model

    def Eval(self, request: messages.EvalRequest, context) -> messages.EvalResponse:  # pyright: ignore[reportImplicitOverride]
        response = messages.EvalResponse()
        accuracy, loss, guesses = self.model.test()
        response.accuracy = accuracy
        response.loss = loss
        response.guesses.update(guesses)
        response.success = True
        if request.path:
            model_path = f"{request.path}.pt"
            Path(model_path).parent.mkdir(parents=True, exist_ok=True)
            save(self.model.export_model_weights(), model_path)
            logging.info(f"Saved model checkpoint to {model_path}")
        return response

class ImportWeightsService(ipc.ImportWeightsServicer):
    def __init__(self, model: Model) -> None:
        super().__init__()
        self.model: Model = model

    def ImportWeights(self, request: messages.ImportRequest, context) -> messages.ImportResponse:  # pyright: ignore[reportImplicitOverride]
        response = messages.ImportResponse()
        try:
            weights = load(BytesIO(request.weights))
            self.model.import_model_weights(
                weights,
                request.weight_ratio
            )
            response.success = True
        except Exception as e:
            response.success = False
            response.error_message = str(e)
            logging.error(f"Error importing weights: {e}")
        return response


class ExportWeightsService(ipc.ExportWeightsServicer):
    def __init__(self, model: Model) -> None:
        super().__init__()
        self.model: Model = model

    def ExportWeights(self, request: messages.ExportRequest, context) -> messages.ExportResponse:  # pyright: ignore[reportImplicitOverride]
        response = messages.ExportResponse()
        weights_buffer = BytesIO()
        save(self.model.export_model_weights(),
                    weights_buffer)
        response.success = True
        response.weights = weights_buffer.getvalue()
        return response
