from fastapi import FastAPI, File, UploadFile
from fastapi.responses import JSONResponse
from PIL import Image
import io
import threading
import queue
from ultralytics import YOLO

app = FastAPI(title="YOLO Service")

model = YOLO("model/best.pt")

frame_queue = queue.Queue()
result_queue = queue.Queue()

def yolo_worker():
    while True:
        item = frame_queue.get()
        if item is None:  
            break
        frame_id, img = item
        results = model(img)
        result = results[0].summary() 
        result_queue.put((frame_id, result))
        frame_queue.task_done()


threading.Thread(target=yolo_worker, daemon=True).start()

frame_counter = 0
frame_lock = threading.Lock()

@app.post("/detect_summary")
async def detect_summary(file: UploadFile = File(...)):
    global frame_counter
    try:
        contents = await file.read()
        img = Image.open(io.BytesIO(contents)).convert("RGB")

        with frame_lock:
            frame_id = frame_counter
            frame_counter += 1

        frame_queue.put((frame_id, img))

        while True:
            r_id, summary = result_queue.get()
            if r_id == frame_id:
                result_queue.task_done()
                break
            else:
            
                result_queue.put((r_id, summary))

        return JSONResponse(content={"frame_id": frame_id, "summary": summary})

    except Exception as e:
        return JSONResponse(content={"error": str(e)}, status_code=500)
    
@app.get("/health")
async def health():
    return JSONResponse(content={"status": "ok"})
