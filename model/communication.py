import logging, io, os, socket

import betterproto
import torch

from training import Model
import lib.model as pb

class ModelServer:
    def __init__(self, model: Model, socket_path: str):
        self.model = model
        self.socket_path = socket_path

    def start(self):
        """Start the model server and handle connections."""
        if os.path.exists(self.socket_path):
            os.unlink(self.socket_path)

        server = socket.socket(socket.AF_UNIX, socket.SOCK_STREAM)
        server.bind(self.socket_path)
        server.listen(1)
        logging.info(f"Listening on {self.socket_path}")

        while True:
            conn, addr = server.accept()
            try:
                self._handle_connection(conn)
            finally:
                conn.close()

    def _ack(self, conn: socket.socket):
        ack = pb.Ack()
        conn.send(bytes(ack))

    def _handle_connection(self, conn: socket.socket):
        """Handle a single connection."""
        while True:
            logging.info("Waiting for command.")
            # Read message length (varint)
            buf = []
            while True:
                buf.append(conn.recv(1))
                try:
                    msg_len, pos = betterproto.decode_varint(b''.join(buf), 0)
                    break
                except IndexError:
                    continue

            # Read the message
            data = conn.recv(msg_len)
            if not data:
                break
            logging.info(f"Got message of length {len(data)}")

            # Parse request
            request = pb.ModelRequest().parse(data)

            logging.info(f"Message: {request}")
            # Create response

            response = pb.ModelResponse(success=False)
            # Handle request
            t, values = betterproto.which_one_of(request, "request")
            match t:
                case "export_weights":
                    self._ack(conn)
                    weights_buffer = io.BytesIO()
                    torch.save(self.model.export_model_weights(), weights_buffer)
                    response.success = True
                    response.weights = weights_buffer.getvalue()

                case "import_weights":
                    self._ack(conn)
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
                    self._ack(conn)
                    avg_loss = self.model.train()
                    response.loss = avg_loss
                    response.success = True

                case "eval":
                    self._ack(conn)
                    accuracy, loss = self.model.test()
                    response.accuracy = accuracy
                    response.loss = loss
                    response.success = True

            # Send response with length prefix
            response_data = bytes(response)
            conn.send(betterproto.encode_varint(len(response_data)))
            conn.send(response_data)
