import logging
import io
import os
import socket
import pathlib

import betterproto
import torch

from training import Model
import lib.model as pb


class ModelServer:
    def __init__(self, model: Model, socket_path: str):
        self.model = model
        self.socket_path = socket_path
        self.conn: socket.socket | None = None

    def start(self):
        """Start the model server and handle connections."""
        if os.path.exists(self.socket_path):
            os.unlink(self.socket_path)

        server = socket.socket(socket.AF_UNIX, socket.SOCK_STREAM)
        server.bind(self.socket_path)
        server.listen(1)
        logging.info(f"Listening on {self.socket_path}")

        while True:
            self.conn, addr = server.accept()
            try:
                self._handle_connection()
            finally:
                self.conn.close()

    def _handle_connection(self):
        """Handle a single connection."""
        assert self.conn is not None

        def _ack():
            assert self.conn is not None
            self.conn.sendall(betterproto.encode_varint(42))

        while True:
            logging.info("Waiting for command")
            # Read message length (varint)
            buf = []
            while True:
                buf.append(self.conn.recv(1))
                try:
                    msg_len, pos = betterproto.decode_varint(b''.join(buf), 0)
                    break
                except IndexError:
                    continue

            # Read the message
            data = b""
            while len(data) < msg_len:
                data += self.conn.recv(msg_len)
                if not data:
                    # Client has disconnected
                    raise Exception("Client disconnected")

            # Parse request
            request = pb.ModelRequest().parse(data)

            # Handle request
            response = pb.ModelResponse(success=False)
            t, values = betterproto.which_one_of(request, "request")
            match t:
                case "export_weights":
                    _ack()
                    logging.info("Exporting weights")
                    weights_buffer = io.BytesIO()
                    torch.save(self.model.export_model_weights(),
                               weights_buffer)
                    response.success = True
                    response.weights = weights_buffer.getvalue()

                case "import_weights":
                    _ack()
                    logging.info("Importing weights")
                    try:
                        weights = torch.load(io.BytesIO(values.weights))
                        self.model.import_model_weights(
                            weights,
                            values.weight_ratio
                        )
                        response.success = True
                    except Exception as e:
                        response.success = False
                        response.error_message = str(e)

                case "train":
                    _ack()
                    logging.info("Training model")
                    avg_loss = self.model.train()
                    response.loss = avg_loss
                    response.success = True

                case "eval":
                    _ack()
                    logging.info("Evaluating model")
                    accuracy, loss = self.model.test()
                    response.accuracy = accuracy
                    response.loss = loss
                    response.success = True
                    # Save model weights to disk after evaluation
                    if values.path:
                        model_path = f"{values.path}.pt"
                        pathlib.Path(model_path).parent.mkdir(parents=True, exist_ok=True)
                        torch.save(self.model.export_model_weights(), model_path)
                        logging.info(f"Saved model checkpoint to {model_path}")

            # Send response with length prefix
            response_data = bytes(response)
            self.conn.send(betterproto.encode_varint(len(response_data)))
            self.conn.send(response_data)
