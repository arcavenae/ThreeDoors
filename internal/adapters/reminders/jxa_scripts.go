package reminders

import (
	"fmt"
	"strings"
)

// escapeJXA escapes a string for safe embedding in a JXA script literal.
// Prevents injection by escaping backslashes, double quotes, and newlines.
func escapeJXA(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	s = strings.ReplaceAll(s, "\n", `\n`)
	s = strings.ReplaceAll(s, "\r", `\r`)
	return s
}

// scriptReadReminders returns a JXA script that reads incomplete reminders
// from the specified list as a JSON array.
func scriptReadReminders(listName string) string {
	return fmt.Sprintf(`(function() {
  var app = Application("Reminders");
  var list = app.lists.byName("%s");
  var items = list.reminders.whose({completed: false})();
  var result = [];
  for (var i = 0; i < items.length; i++) {
    var r = items[i];
    result.push({
      id: r.id(),
      name: r.name(),
      body: r.body() || "",
      dueDate: r.dueDate() ? r.dueDate().toISOString() : null,
      priority: r.priority(),
      completed: r.completed(),
      flagged: r.flagged(),
      creationDate: r.creationDate().toISOString(),
      modificationDate: r.modificationDate().toISOString()
    });
  }
  return JSON.stringify(result);
})()`, escapeJXA(listName))
}

// scriptReadLists returns a JXA script that reads all reminder list names
// as a JSON array of strings.
func scriptReadLists() string {
	return `(function() {
  var app = Application("Reminders");
  var lists = app.lists();
  var names = [];
  for (var i = 0; i < lists.length; i++) {
    names.push(lists[i].name());
  }
  return JSON.stringify(names);
})()`
}

// scriptCompleteReminder returns a JXA script that marks a reminder as
// completed by its ID.
func scriptCompleteReminder(reminderID string) string {
	return fmt.Sprintf(`(function() {
  var app = Application("Reminders");
  var lists = app.lists();
  for (var i = 0; i < lists.length; i++) {
    var reminders = lists[i].reminders();
    for (var j = 0; j < reminders.length; j++) {
      if (reminders[j].id() === "%s") {
        reminders[j].completed = true;
        return JSON.stringify({success: true});
      }
    }
  }
  return JSON.stringify({success: false, error: "reminder not found"});
})()`, escapeJXA(reminderID))
}

// scriptCreateReminder returns a JXA script that creates a new reminder
// in the specified list with the given properties.
func scriptCreateReminder(name, body string, priority int, listName string) string {
	return fmt.Sprintf(`(function() {
  var app = Application("Reminders");
  var list = app.lists.byName("%s");
  var reminder = app.Reminder({
    name: "%s",
    body: "%s",
    priority: %d
  });
  list.reminders.push(reminder);
  return JSON.stringify({success: true, id: reminder.id()});
})()`, escapeJXA(listName), escapeJXA(name), escapeJXA(body), priority)
}

// scriptUpdateReminder returns a JXA script that updates an existing reminder's
// name, body, and priority by its ID.
func scriptUpdateReminder(reminderID, name, body string, priority int) string {
	return fmt.Sprintf(`(function() {
  var app = Application("Reminders");
  var lists = app.lists();
  for (var i = 0; i < lists.length; i++) {
    var reminders = lists[i].reminders();
    for (var j = 0; j < reminders.length; j++) {
      if (reminders[j].id() === "%s") {
        reminders[j].name = "%s";
        reminders[j].body = "%s";
        reminders[j].priority = %d;
        return JSON.stringify({success: true});
      }
    }
  }
  return JSON.stringify({success: false, error: "reminder not found"});
})()`, escapeJXA(reminderID), escapeJXA(name), escapeJXA(body), priority)
}

// scriptDeleteReminder returns a JXA script that deletes a reminder by its ID.
func scriptDeleteReminder(reminderID string) string {
	return fmt.Sprintf(`(function() {
  var app = Application("Reminders");
  var lists = app.lists();
  for (var i = 0; i < lists.length; i++) {
    var reminders = lists[i].reminders();
    for (var j = 0; j < reminders.length; j++) {
      if (reminders[j].id() === "%s") {
        app.delete(reminders[j]);
        return JSON.stringify({success: true});
      }
    }
  }
  return JSON.stringify({success: false, error: "reminder not found"});
})()`, escapeJXA(reminderID))
}
