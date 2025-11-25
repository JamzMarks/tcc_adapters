from fastapi import FastAPI
from pydantic import BaseModel
from datetime import datetime
from rabbit_publisher import publish_message

app = FastAPI()

class Data(BaseModel):
    confiability: float
    flow: float

class MessageIn(BaseModel):
    deviceId: str
    deviceType: str
    data: Data

@app.post("/publish")
def publish(msg: MessageIn):

    message_out = {
        "deviceId": msg.deviceId,
        "deviceType": msg.deviceType,
        "data": {
            "confiability": msg.data.confiability,
            "flow": msg.data.flow,
        },
        "ts": datetime.utcnow().isoformat() + "Z"
    }

    publish_message(message_out)

    return {
        "status": "published",
        "sent": message_out
    }
