# Test sample: SQL Injection vulnerabilities
# This file contains intentionally vulnerable code for testing

import sqlite3

def get_user_vulnerable(user_id):
    """VULNERABLE: SQL injection via string concatenation"""
    conn = sqlite3.connect('users.db')
    cursor = conn.cursor()
    # Vulnerable - user_id directly concatenated
    query = "SELECT * FROM users WHERE id = " + user_id
    cursor.execute(query)
    return cursor.fetchone()

def search_users_vulnerable(name):
    """VULNERABLE: SQL injection via f-string"""
    conn = sqlite3.connect('users.db')
    cursor = conn.cursor()
    # Vulnerable - f-string with user input
    query = f"SELECT * FROM users WHERE name LIKE '%{name}%'"
    cursor.execute(query)
    return cursor.fetchall()

def login_vulnerable(username, password):
    """VULNERABLE: SQL injection in authentication"""
    conn = sqlite3.connect('users.db')
    cursor = conn.cursor()
    # Vulnerable - authentication bypass possible
    query = f"SELECT * FROM users WHERE username = '{username}' AND password = '{password}'"
    cursor.execute(query)
    return cursor.fetchone() is not None

# SECURE versions for comparison

def get_user_secure(user_id):
    """SECURE: Parameterized query"""
    conn = sqlite3.connect('users.db')
    cursor = conn.cursor()
    cursor.execute("SELECT * FROM users WHERE id = ?", (user_id,))
    return cursor.fetchone()

def search_users_secure(name):
    """SECURE: Parameterized query with LIKE"""
    conn = sqlite3.connect('users.db')
    cursor = conn.cursor()
    cursor.execute("SELECT * FROM users WHERE name LIKE ?", (f"%{name}%",))
    return cursor.fetchall()
