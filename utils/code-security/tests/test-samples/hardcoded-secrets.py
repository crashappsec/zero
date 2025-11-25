# Test sample: Hardcoded secrets and credentials
# This file contains intentionally vulnerable code for testing

import requests
import boto3

# VULNERABLE: Hardcoded API keys
API_KEY = "sk-1234567890abcdef1234567890abcdef"
STRIPE_KEY = "sk_live_4eC39HqLyjWDarjtT1zdp7dc"
GITHUB_TOKEN = "ghp_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"

# VULNERABLE: Hardcoded database credentials
DB_HOST = "localhost"
DB_USER = "admin"
DB_PASSWORD = "super_secret_password_123"

# VULNERABLE: Hardcoded AWS credentials
AWS_ACCESS_KEY = "AKIAIOSFODNN7EXAMPLE"
AWS_SECRET_KEY = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"

# VULNERABLE: Hardcoded JWT secret
JWT_SECRET = "my-super-secret-jwt-signing-key-do-not-share"

def connect_to_api():
    """VULNERABLE: Using hardcoded API key"""
    headers = {"Authorization": f"Bearer {API_KEY}"}
    return requests.get("https://api.example.com/data", headers=headers)

def connect_to_database():
    """VULNERABLE: Using hardcoded credentials"""
    connection_string = f"mysql://{DB_USER}:{DB_PASSWORD}@{DB_HOST}/mydb"
    return connection_string

def connect_to_aws():
    """VULNERABLE: Using hardcoded AWS credentials"""
    client = boto3.client(
        's3',
        aws_access_key_id=AWS_ACCESS_KEY,
        aws_secret_access_key=AWS_SECRET_KEY
    )
    return client

# SECURE versions for comparison

import os

def connect_to_api_secure():
    """SECURE: Using environment variable"""
    api_key = os.environ.get("API_KEY")
    if not api_key:
        raise ValueError("API_KEY environment variable not set")
    headers = {"Authorization": f"Bearer {api_key}"}
    return requests.get("https://api.example.com/data", headers=headers)

def connect_to_database_secure():
    """SECURE: Using environment variables"""
    db_user = os.environ.get("DB_USER")
    db_password = os.environ.get("DB_PASSWORD")
    db_host = os.environ.get("DB_HOST", "localhost")
    return f"mysql://{db_user}:{db_password}@{db_host}/mydb"
