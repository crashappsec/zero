# Test file: Python/Flask detection patterns
# Should detect: Python, Flask, SQLAlchemy, possibly JWT patterns

from flask import Flask, request, jsonify, render_template
from flask_sqlalchemy import SQLAlchemy
from flask_jwt_extended import JWTManager, jwt_required, create_access_token
import os

app = Flask(__name__)
app.config['SQLALCHEMY_DATABASE_URI'] = os.environ.get('DATABASE_URL')
app.config['JWT_SECRET_KEY'] = os.environ.get('JWT_SECRET')

db = SQLAlchemy(app)
jwt = JWTManager(app)


class User(db.Model):
    id = db.Column(db.Integer, primary_key=True)
    username = db.Column(db.String(80), unique=True, nullable=False)
    email = db.Column(db.String(120), unique=True, nullable=False)

    def to_dict(self):
        return {'id': self.id, 'username': self.username, 'email': self.email}


@app.route('/api/login', methods=['POST'])
def login():
    username = request.json.get('username')
    password = request.json.get('password')

    user = User.query.filter_by(username=username).first()
    if user and check_password(password):
        token = create_access_token(identity=user.id)
        return jsonify(access_token=token)

    return jsonify(error='Invalid credentials'), 401


@app.route('/api/users', methods=['GET'])
@jwt_required()
def get_users():
    users = User.query.all()
    return jsonify([u.to_dict() for u in users])


@app.route('/api/users', methods=['POST'])
def create_user():
    data = request.json
    user = User(username=data['username'], email=data['email'])
    db.session.add(user)
    db.session.commit()
    return jsonify(user.to_dict()), 201


if __name__ == '__main__':
    app.run(debug=True)
