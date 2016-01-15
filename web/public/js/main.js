function doneTyping (t) {
  $.ajax({
    method: "POST",
    url: "/elasticbook/suggest",
    data: { term: t }
  })
  .done(function() {
    console.log( "sent request", t );
  })
  .fail(function() {
    console.log( "error" );
  })
  .always(function() {
    console.log("done...")
  });
}

$(document).on('keydown', 'form input[type=text]', function(e) {
  var typingTimer;
  var doneTypingInterval = 2000;
  var $input = $(this);

  if (($input.data('suggest') == true)) {
    $input.on('keyup', function () {
      clearTimeout(typingTimer);
      var term = $input.val()
      typingTimer = setTimeout(doneTyping(term), doneTypingInterval);
    });

    $input.on('keydown', function () {
      clearTimeout(typingTimer);
    });
  };
});
