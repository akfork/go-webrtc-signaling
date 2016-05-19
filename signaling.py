# -*- coding: utf-8  -*-



import json
import jwt
from urlparse import urlparse
from gevent import monkey
monkey.patch_all()


from flask import Flask, app, render_template
from werkzeug.debug import DebuggedApplication


from geventwebsocket import WebSocketServer, WebSocketApplication,\
            Resource


flask_app = Flask(__name__)
flask_app.debug = True




class Message(object):

    def __init__(self):

        self.user_id = ''
        self.type = ''
        self.room = 'default'
        self.message = None


class SignalingApplication(WebSocketApplication):

    def on_open(self):

        print 'connection opned '

        print 'client_address ',self.ws.handler.client_address

        token = ''
        room = ''
        user_id = ''

        print self.ws.handler.server.clients.values()

        print self.ws.environ['PATH_INFO']

        path = self.ws.environ['PATH_INFO']

        token = path.split('/')[-1]

        print token

        try:
            _data = jwt.decode(token,'signaling')
        except:

            print 'auth error'
            self.ws.close()
            return


        current_client = self.ws.handler.active_client


        setattr(current_client,'room',_data['room'])
        setattr(current_client,'user_id',_data['user_id'])


        self.peer_connected(current_client.room,current_client.user_id)

        self.send_peers(current_client.room,current_client.user_id)


    def on_message(self,message):

        if self.ws.closed:
            return

        message = json.loads(message)

        print  'message',message

        current_client = self.ws.handler.active_client

        if message['eventName'] == 'offer':

            _m = self.format_message(current_client,'offer',message)

            self.send_one(current_client.room,message['targetUserId'],_m)


        if message['eventName'] == 'answer':

            _m = self.format_message(current_client,'answer',message)

            self.send_one(current_client.room,message['targetUserId'],_m)


        if message['eventName'] == 'ice':

            _m = self.format_message(current_client,'ice',message)

            self.send_one(current_client.room,message['targetUserId'],_m)


        if message['eventName'] == 'bye':

            _m = self.format_message(current_client,'bye',message)

            self.send_one(current_client.room,message['targetUserId'],_m)




    def on_close(self,reason):

        print 'connection closed',reason

        # some handler
        current_client = self.ws.handler.active_client

        self.peer_removed(current_client.room,current_client.user_id)


    def format_message(self,client,_type,message):

        _m = {
                'user_id':client.user_id,
                'type':_type,
                'room':client.room,
                'message':message
                }

        return _m


    def broadcast(self,room,user_id,message):

        # 向房间发送  除了user_id
        for client in self.ws.handler.server.clients.values():
            if client.room == room and client.user_id != user_id:
                client.ws.send(json.dumps(message))


    def send_one(self,room,user_id,message):

        #  向房间的 user_id 发送消息
        for client in self.ws.handler.server.clients.values():
            if client.room == room and client.user_id == user_id:
                client.ws.send(json.dumps(message))



    def send_peers(self,room,user_id):

        users = []
        for client in self.ws.handler.server.clients.values():
            if client.room == room:
                users.append(getattr(client,'user_id','anonymous'))

        message = {
                'user_id':'system',
                'type':'peers',
                'room':room,
                'message':{'users':users}
                }

        self.send_one(room,user_id,message)


    def peer_connected(self,room,user_id):

        message = {
                'user_id':user_id,
                'type':'peer_connected',
                'room': room,
                'message':'kidding'
                }

        self.broadcast(room,user_id,message)


    def peer_removed(self,room,user_id):

        message = {
                'user_id':user_id,
                'type':'peer_removed',
                'room': room,
                'message':'kidding'
                }

        self.broadcast(room,user_id,message)




@flask_app.route('/token/<room>/<user_id>')
def token(room,user_id):

    token = jwt.encode({'room':room,
                        'user_id':user_id},'signaling')
    return token,200



WebSocketServer(
        ('0.0.0.0', 8000),
        Resource([('^/ws.*',SignalingApplication),
                ('^/.*',DebuggedApplication(flask_app))])
).serve_forever()
