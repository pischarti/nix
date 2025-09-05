import unittest
from fastapi.testclient import TestClient

from py.fastapi.helloworld import app


class HelloWorldApiTests(unittest.TestCase):
    def setUp(self) -> None:
        self.client = TestClient(app)

    def test_read_root(self) -> None:
        res = self.client.get("/")
        self.assertEqual(res.status_code, 200)
        self.assertEqual(res.json(), {"message": "Hello, World!"})

    def test_greet(self) -> None:
        res = self.client.get("/hello/Ada")
        self.assertEqual(res.status_code, 200)
        self.assertEqual(res.json(), {"message": "Hello, Ada!"})

    def test_create_item(self) -> None:
        payload = {"name": "book", "description": "novel", "price": 12.5}
        res = self.client.post("/items/", json=payload)
        self.assertEqual(res.status_code, 200)
        self.assertEqual(res.json()["name"], "book")
        self.assertEqual(res.json()["price"], 12.5)

    def test_signup_background_task(self) -> None:
        res = self.client.post("/signup", params={"email": "user@example.com"})
        self.assertEqual(res.status_code, 200)
        self.assertIn("signed up", res.json()["message"]) 

    def test_protected_requires_api_key(self) -> None:
        res = self.client.get("/protected")
        self.assertEqual(res.status_code, 401)

    def test_protected_with_api_key(self) -> None:
        res = self.client.get("/protected", headers={"X-API-Key": "SECRET_API_KEY"})
        self.assertEqual(res.status_code, 200)
        self.assertEqual(res.json(), {"message": "Success! You are authenticated."})


if __name__ == "__main__":
    unittest.main()


