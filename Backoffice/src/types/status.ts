export function getStatusLabel(status?: string): string {
  if (!status) return "";

  switch (status) {
    case "new":
      return "Новые";
    case "resolved":
      return "Решённые";
    case "open":
      return "Открытые";
    case "closed":
      return "Решённые";
    case "partially_closed":
    case "partially_resolved":
      return "Частично решённые";
    default:
      return status;
  }
}

export function getRequestStatusLabel(status?: string): string {
  if (!status) return "действие";

  switch (status) {
    case "open":
      return "Открыто";
    case "partially_closed":
    case "partially_resolved":
      return "Частичное закрытие";
    case "closed":
      return "Закрыто";
    default:
      return status;
  }
}


