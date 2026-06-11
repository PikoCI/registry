// namespace.js - App namespace and utilities
window.PkReg = window.PkReg || {};

PkReg.Theme = {
  current: function() {
    return localStorage.getItem('pikoci-registry-theme') || 'light';
  },
  toggle: function() {
    var next = this.current() === 'light' ? 'dark' : 'light';
    localStorage.setItem('pikoci-registry-theme', next);
    document.documentElement.setAttribute('data-theme', next);
    return next;
  },
  init: function() {
    document.documentElement.setAttribute('data-theme', this.current());
  }
};

PkReg.Auth = {
  getToken: function() {
    return localStorage.getItem('pikoci-registry-token');
  },
  setToken: function(token) {
    localStorage.setItem('pikoci-registry-token', token);
  },
  getUser: function() {
    var u = localStorage.getItem('pikoci-registry-user');
    return u ? JSON.parse(u) : null;
  },
  setUser: function(user) {
    localStorage.setItem('pikoci-registry-user', JSON.stringify(user));
  },
  isLoggedIn: function() {
    return !!this.getToken() && !!this.getUser();
  },
  logout: function() {
    localStorage.removeItem('pikoci-registry-token');
    localStorage.removeItem('pikoci-registry-user');
  },
  ajaxHeaders: function() {
    var h = { 'Content-Type': 'application/json' };
    var t = this.getToken();
    if (t) { h['Authorization'] = 'Bearer ' + t; }
    return h;
  }
};

// Utility to format kind labels
PkReg.formatKind = function(kind) {
  if (!kind) return '';
  return kind.replace(/_/g, ' ').replace(/\b\w/g, function(c) { return c.toUpperCase(); });
};

// Utility to format dates
PkReg.formatDate = function(dateStr) {
  if (!dateStr) return '';
  var d = new Date(dateStr);
  return d.toLocaleDateString('en-US', { year: 'numeric', month: 'short', day: 'numeric' });
};

// Kind icon mapping
PkReg.kindIcon = function(kind) {
  var icons = {
    'resource_type': 'bi-box',
    'runner_type': 'bi-play-circle',
    'service_type': 'bi-gear',
    'secret_type': 'bi-shield-lock',
    'notification_type': 'bi-bell'
  };
  return icons[kind] || 'bi-box';
};
