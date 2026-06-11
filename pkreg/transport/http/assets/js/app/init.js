// init.js - App bootstrap
(function() {
  'use strict';

  $(function() {
    // Initialize theme
    PkReg.Theme.init();

    // Create main view
    var mainView = new PkReg.Views.MainView();

    // Create and start router
    PkReg.app = {
      mainView: mainView,
      router: new PkReg.Router({ mainView: mainView })
    };

    // Start Backbone history with pushState
    Backbone.history.start({ pushState: true });

    // Intercept link clicks for client-side routing
    $(document).on('click', 'a[href^="/"]', function(e) {
      if (e.ctrlKey || e.metaKey || e.shiftKey) return;
      e.preventDefault();
      var path = $(this).attr('href').replace(/^\//, '');
      Backbone.history.navigate(path, { trigger: true });
    });

    // Also handle hash-based links starting with #
    $(document).on('click', 'a[href^="#"]', function(e) {
      var href = $(this).attr('href');
      // Skip if it's just "#" or starts with "#/" (handled by Backbone)
      if (href === '#' || href === '#/') return;
      // For anchors like #plugins/ns/name, navigate
      if (href.length > 1 && href.charAt(0) === '#') {
        e.preventDefault();
        var path = href.substring(1);
        Backbone.history.navigate(path, { trigger: true });
      }
    });
  });

})();
