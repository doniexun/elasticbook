function doneTyping () {
  console.log("done...")
}

$(document).on('keydown', 'form input[type=text]', function(e) {
  var typingTimer;
  var doneTypingInterval = 5000;
  var $input = $(this);

  if (($input.data('suggest') == true)) {
    $input.on('keyup', function () {
      clearTimeout(typingTimer);
      typingTimer = setTimeout(doneTyping, doneTypingInterval);
    });

    $input.on('keydown', function () {
      clearTimeout(typingTimer);
    });
  };
});
