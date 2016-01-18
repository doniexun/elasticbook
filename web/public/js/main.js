function doneTyping (t) {
  $.ajax({
    method: "POST",
    url: "/elasticbook/suggest",
    data: { term: t }
  })
  .done(function(res) {
    console.log("received suggestions");

    var suggestions = [];
    var options = res.completion[0]["options"];
    options.forEach(function(v) {
      suggestions.push(v['text'])
    })
    console.log(suggestions);
    $("form input[type=text][data-suggest=true]").autocomplete({
      source: suggestions });
  })
  .fail(function() {
    console.log("error");
  })
  .always(function() {
    console.log("done...")
  });
}

// new Awesomplete(
//   $("form input[type=text][data-suggest=true]").get(0),
//   {
//     minChars: 3,
//     maxItems: 15,
//     list: suggestions
//   });

// var awesomplete = new Awesomplete(
//   input,
//   {
//     minChars: 3,
//     maxItems: 15
//   });
// list = [];
// awesomplete.list = list;

$(document).ready(function() {
  if (window.location.href.indexOf("/search") == -1) {
    $("form input[type=text][data-suggest=true]").get(0).focus();
  }
});

$(document).on('keydown', 'form input[type=text]', function(e) {
  var typingTimer;
  var doneTypingInterval = 2000;
  var input = $(this);

  if ((input.data('suggest') == true)) {
    input.on('keyup', function () {
      clearTimeout(typingTimer);
      var term = input.val()
      typingTimer = setTimeout(doneTyping(term), doneTypingInterval);
    });

    input.on('keydown', function () {
      clearTimeout(typingTimer);
    });
  };
});

// new Awesomplete($('input[type="email"]'), {
//   list: ["@aol.com", "@att.net", "@comcast.net", "@facebook.com",
//          "@gmail.com", "@gmx.com", "@googlemail.com", "@google.com",
//          "@hotmail.com", "@hotmail.co.uk", "@mac.com", "@me.com",
//          "@mail.com", "@msn.com", "@live.com", "@sbcglobal.net",
//          "@verizon.net", "@yahoo.com", "@yahoo.co.uk"],
//   item: function(text, input){
//     var newText = input.slice(0, input.indexOf("@")) + text;
//     return Awesomplete.$.create("li", {
//         innerHTML: newText.replace(RegExp(input.trim(), "gi"), "<mark>$&</mark>"),
//         "aria-selected": "false"
//     });
//     },
//   filter: function(text, input){
//     return RegExp("^" + Awesomplete.$.regExpEscape(input.replace(/^.+?(?=@)/, ''), "i")).test(text);
//     }
// });

// var ajax
// ajax = new XMLHttpRequest()
// ajax.open("GET", "https://restcountries.eu/rest/v1/lang/fr", true)
// ajax.onload = function() {
//   {
//       var list = JSON.parse(ajax.responseText).map(function(i) {
//       return i.name;
//       })
//       new Awesomplete(document.querySelector("#ajax-example input"),{list:list});
//   }
// }
// ajax.send();
