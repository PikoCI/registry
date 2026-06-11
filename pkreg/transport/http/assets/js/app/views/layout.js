// views/layout.js - MainView and HeaderView
(function() {
  'use strict';

  PkReg.Views = PkReg.Views || {};

  // HeaderView - Navbar with search, theme toggle, user menu
  PkReg.Views.HeaderView = Backbone.View.extend({
    template: _.template($('#tmpl-header').html()),

    events: {
      'click .theme-toggle': 'toggleTheme',
      'submit .navbar-search-form': 'onSearch',
      'click .logout-btn': 'onLogout'
    },

    initialize: function() {
      this.render();
    },

    render: function() {
      var user = PkReg.Auth.getUser();
      var isLoggedIn = PkReg.Auth.isLoggedIn();
      // If token exists but user data is missing, clear stale auth state
      if (isLoggedIn && !user) {
        PkReg.Auth.logout();
        isLoggedIn = false;
      }
      var theme = PkReg.Theme.current();
      this.$el.html(this.template({
        user: user,
        isLoggedIn: isLoggedIn,
        theme: theme
      }));
      return this;
    },

    toggleTheme: function(e) {
      e.preventDefault();
      var theme = PkReg.Theme.toggle();
      this.$('.theme-toggle i').attr('class', theme === 'dark' ? 'bi bi-sun' : 'bi bi-moon');
    },

    onSearch: function(e) {
      e.preventDefault();
      var q = this.$('.navbar-search-input').val().trim();
      if (q) {
        Backbone.history.navigate('search?q=' + encodeURIComponent(q), { trigger: true });
      } else {
        Backbone.history.navigate('', { trigger: true });
      }
    },

    onLogout: function(e) {
      e.preventDefault();
      PkReg.Auth.logout();
      this.render();
      Backbone.history.navigate('', { trigger: true });
    }
  });

  // MainView - Renders header + content area
  PkReg.Views.MainView = Backbone.View.extend({
    el: '#app',

    initialize: function() {
      this.$el.html('<div id="header-region"></div><main class="container py-4" id="content-region"></main><footer class="pkreg-footer"><div class="container">PikoCI Registry</div></footer>');
      this.header = new PkReg.Views.HeaderView({ el: '#header-region' });
    },

    showContent: function(view) {
      if (this.currentView && this.currentView.remove) {
        this.currentView.remove();
      }
      this.currentView = view;
      $('#content-region').html('').append(view.render().el);
    },

    refreshHeader: function() {
      this.header.render();
    }
  });

})();
