// Test sample: Cross-Site Scripting (XSS) vulnerabilities
// This file contains intentionally vulnerable code for testing

// VULNERABLE: innerHTML with user input
function displayUserMessage(message) {
  const container = document.getElementById('messages');
  // Vulnerable - XSS via innerHTML
  container.innerHTML = '<div class="message">' + message + '</div>';
}

// VULNERABLE: document.write with user input
function showWelcome(username) {
  // Vulnerable - XSS via document.write
  document.write('<h1>Welcome, ' + username + '!</h1>');
}

// VULNERABLE: jQuery .html() with user input
function updateContent(content) {
  // Vulnerable - XSS via jQuery html()
  $('#content').html(content);
}

// VULNERABLE: eval with user input
function processData(userCode) {
  // Vulnerable - arbitrary code execution
  eval(userCode);
}

// VULNERABLE: setAttribute for event handlers
function createButton(onclick) {
  const btn = document.createElement('button');
  // Vulnerable - XSS via onclick attribute
  btn.setAttribute('onclick', onclick);
  document.body.appendChild(btn);
}

// VULNERABLE: href with user input (javascript: protocol)
function createLink(url, text) {
  const link = document.createElement('a');
  // Vulnerable - javascript: protocol XSS
  link.href = url;
  link.textContent = text;
  document.body.appendChild(link);
}

// SECURE versions for comparison

// SECURE: textContent instead of innerHTML
function displayUserMessageSecure(message) {
  const container = document.getElementById('messages');
  const div = document.createElement('div');
  div.className = 'message';
  div.textContent = message;  // Safe - automatically escaped
  container.appendChild(div);
}

// SECURE: Using DOMPurify for sanitization
function updateContentSecure(content) {
  // Safe - sanitized with DOMPurify
  $('#content').html(DOMPurify.sanitize(content));
}

// SECURE: Event listener instead of onclick attribute
function createButtonSecure(callback) {
  const btn = document.createElement('button');
  btn.addEventListener('click', callback);  // Safe - no string eval
  document.body.appendChild(btn);
}

// SECURE: URL validation
function createLinkSecure(url, text) {
  const link = document.createElement('a');
  // Validate URL protocol
  const parsed = new URL(url, window.location.origin);
  if (parsed.protocol === 'http:' || parsed.protocol === 'https:') {
    link.href = url;
  } else {
    link.href = '#';  // Safe fallback
  }
  link.textContent = text;
  document.body.appendChild(link);
}
