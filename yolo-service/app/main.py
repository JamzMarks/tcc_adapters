from fastapi import FastAPI, File, UploadFile, Query
from fastapi.responses import JSONResponse, StreamingResponse
from fastapi.middleware.cors import CORSMiddleware
from PIL import Image
import io
import threading
import queue
from ultralytics import YOLO
import numpy as np
import requests

app = FastAPI(title="YOLO Service")

# --- CORS ---
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],       
    allow_credentials=True,
    allow_methods=["*"],          
    allow_headers=["*"],          
)
# -------------

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


@app.post("/test")
async def test(file: UploadFile = File(...)):
    try:
        content = await file.read()
        img = Image.open(io.BytesIO(content)).convert("RGB")

        img_np = np.array(img)
        results = model(img_np)

        annotated_img = results[0].plot()
        annotated_pil = Image.fromarray(annotated_img)

        buf = io.BytesIO()
        annotated_pil.save(buf, format="JPEG")
        buf.seek(0)

        return StreamingResponse(buf, media_type="image/jpeg")

    except Exception as e:
        return JSONResponse(content={"error": str(e)}, status_code=500)
    
@app.post("/testcamera")
async def testcamera(ip: str = Query(...), port: int = Query(...)):
    """
    Acessa o snapshot de uma câmera pelo IP/porta, processa com YOLO e retorna a imagem anotada.
    Exemplo: POST http://localhost:7676/testcamera?ip=192.168.0.105&port=8080
    """
    try:
        # Monta a URL do snapshot
        snapshot_url = f"http://{ip}:{port}/snapshot.jpg"
        print(f"[INFO] Capturando snapshot de: {snapshot_url}")

        # Faz a requisição GET à câmera
        response = requests.get(snapshot_url, timeout=5)

        if response.status_code != 200:
            return JSONResponse(
                content={"error": f"Falha ao obter snapshot ({response.status_code})"},
                status_code=500,
            )

        # Converte o conteúdo em imagem PIL
        img = Image.open(io.BytesIO(response.content)).convert("RGB")

        # Converte para numpy array
        img_np = np.array(img)

        # Processa com YOLO
        results = model(img_np)
        annotated_img = results[0].plot()

        # Converte de volta para imagem PIL
        annotated_pil = Image.fromarray(annotated_img)

        # Prepara buffer para enviar
        buf = io.BytesIO()
        annotated_pil.save(buf, format="JPEG")
        buf.seek(0)

        # Retorna imagem processada
        return StreamingResponse(buf, media_type="image/jpeg")

    except requests.exceptions.RequestException as e:
        return JSONResponse(
            content={"error": f"Erro de conexão com a câmera: {str(e)}"},
            status_code=500,
        )

    except Exception as e:
        return JSONResponse(content={"error": str(e)}, status_code=500)


@app.get("/health")
async def health():
    return JSONResponse(content={"status": "ok"})
