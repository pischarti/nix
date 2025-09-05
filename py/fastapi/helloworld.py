from fastapi import Depends, FastAPI, HTTPException, Header, BackgroundTasks
from typing import Dict, Annotated
from pydantic import BaseModel

app = FastAPI(title="HelloWorld API")

# 1. Define a Pydantic model for the item data
class Item(BaseModel):
    name: str
    description: str | None = None
    price: float
    tax: float | None = None

# 2. Use the model as a parameter in a POST endpoint
@app.post("/items/")
def create_item(item: Item):
    """
    Creates a new item with data from the request body.
    """
    return item

@app.get("/")
def read_root() -> Dict[str, str]:
    return {"message": "Hello, World!"}

@app.get("/hello/{name}")
def greet(name: str) -> Dict[str, str]:
    return {"message": f"Hello, {name}!"}

# A function to simulate a long-running task, like sending an email
def send_welcome_email(email: str):
    """Simulates sending a welcome email."""
    print(f"Starting email task for {email}...")
    time.sleep(5)  # Simulate network delay
    print(f"Email sent to {email} successfully!")

@app.post("/signup")
def user_signup(email: str, background_tasks: BackgroundTasks):
    """
    Signs up a user and sends a welcome email in the background.
    """
    # Add the email task to the background_tasks object
    background_tasks.add_task(send_welcome_email, email)
    return {"message": f"User {email} signed up. Sending welcome email..."}

# 1. Define a dependency function that requires an API key header
def get_api_key(x_api_key: Annotated[str, Header()]):
    """
    Authenticates requests using the 'X-API-Key' header.
    """
    if x_api_key != "SECRET_API_KEY": # Replace with actual key lookup
        raise HTTPException(status_code=401, detail="Invalid API Key")
    return x_api_key

# 2. Use the dependency in a protected path operation
@app.get("/protected")
def protected_route(api_key: Annotated[str, Depends(get_api_key)]):
    """
    This endpoint requires a valid API key to access.
    """
    return {"message": "Success! You are authenticated."}

# Optional: allow `python py/fastapi/helloworld.py` to run the server.
if __name__ == "__main__":
    import uvicorn

    uvicorn.run("py.fastapi.helloworld:app", host="127.0.0.1", port=8000, reload=True)


