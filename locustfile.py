import random
import string
import time

from locust import HttpUser, task
from locust_plugins.users.socketio import SocketIOUser


class WebsiteUser(HttpUser):
    host = "http://127.0.0.1:8080/"

    @task
    def hello_world(self):
        N = 7
        res = ''.join(random.choices(string.ascii_uppercase +
                                     string.digits, k=N))
        self.client.post("ws/createRoom", json={
            "id": res,
            "Name": "test",
            "StudentId": "test",
            "TeacherId": "123"
        })


class MySocketIOUser(SocketIOUser):
    @task

    def my_task(self):
        self.my_value = None
        self.connect('ws://127.0.0.1:8080/ws/joinRoom/test?userId=123&username=abolfazl')
        self.send('42["subscribe",{"url":"/sport/matches/11995208/draws","sendInitialUpdate": true}]')
