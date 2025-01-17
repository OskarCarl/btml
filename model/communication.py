import socket
import os
import io
import torch
from google.protobuf.internal.decoder import _DecodeVarint32
from google.protobuf.internal.encoder import _EncodeVarint
import peer_model_pb2
from training import export_model_weights, import_model_weights

class ModelServer:
    def __init__(self, model, socket_path):
        self.model = model
        self.socket_path = socket_path

    def start(self):
        """Start the model server and handle connections."""
        if os.path.exists(self.socket_path):
            os.unlink(self.socket_path)

        server = socket.socket(socket.AF_UNIX, socket.SOCK_STREAM)
        server.bind(self.socket_path)
        server.listen(1)
        print(f"Listening on {self.socket_path}")

        while True:
            conn, addr = server.accept()
            try:
                self._handle_connection(conn)
            finally:
                conn.close()

    def _handle_connection(self, conn):
        """Handle a single connection."""
        while True:
            # Read message length (varint)
            buf = []
            while True:
                buf.append(conn.recv(1))
                try:
                    msg_len, pos = _DecodeVarint32(b''.join(buf), 0)
                    break
                except IndexError:
                    continue

            # Read the message
            data = conn.recv(msg_len)
            if not data:
                break

            # Parse request
            request = peer_model_pb2.ModelRequest()
            request.ParseFromString(data)

            # Create response
            response = peer_model_pb2.ModelResponse()

            # Handle request
            if request.HasField('export_weights'):
                weights_buffer = io.BytesIO()
                torch.save(export_model_weights(self.model), weights_buffer)
                response.success = True
                response.weights = weights_buffer.getvalue()

            elif request.HasField('import_weights'):
                try:
                    weights = torch.load(io.BytesIO(request.import_weights.weights))
                    import_model_weights(
                        self.model,
                        weights,
                        request.import_weights.weight_ratio
                    )
                    response.success = True
                except Exception as e:
                    response.success = False
                    response.error_message = str(e)

            # Send response with length prefix
            response_data = response.SerializeToString()
            _EncodeVarint(conn.send, len(response_data))
            conn.send(response_data)
