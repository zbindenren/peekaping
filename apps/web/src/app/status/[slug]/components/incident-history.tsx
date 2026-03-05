import Markdown from "react-markdown";
import rehypeSanitize from "rehype-sanitize";
import { CheckCircle } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import type { IncidentModel } from "@/api/types.gen";
import { useLocalizedTranslation } from "@/hooks/useTranslation";

const borderColors: Record<string, string> = {
  info: "border-l-blue-500",
  warning: "border-l-amber-500",
  danger: "border-l-red-500",
};

const IncidentHistory = ({
  incidents,
}: {
  incidents: IncidentModel[];
}) => {
  const { t } = useLocalizedTranslation();

  if (incidents.length === 0) {
    return null;
  }

  return (
    <div className="mt-8 text-left">
      <h2 className="text-lg font-semibold mb-4">
        {t("status.incidents.history_title")}
      </h2>
      <div className="space-y-3">
        {incidents.map((incident) => (
          <div
            key={incident.id}
            className={`rounded-lg border border-l-4 ${
              borderColors[incident.style ?? "warning"] ??
              borderColors.warning
            } bg-card p-4`}
          >
            <div className="flex items-start justify-between gap-2">
              <h3 className="font-semibold">{incident.title}</h3>
              {!incident.active && (
                <Badge
                  variant="outline"
                  className="shrink-0 text-green-600 border-green-300"
                >
                  <CheckCircle className="h-3 w-3 mr-1" />
                  {t("status.incidents.resolved")}
                </Badge>
              )}
            </div>
            {incident.content && (
              <div className="mt-1 text-sm prose prose-sm dark:prose-invert max-w-none">
                <Markdown rehypePlugins={[rehypeSanitize]}>
                  {incident.content}
                </Markdown>
              </div>
            )}
            <div className="mt-2 text-xs text-muted-foreground flex gap-4">
              {incident.created_at && (
                <span>
                  {t("common.created")}:{" "}
                  {new Date(incident.created_at).toLocaleString()}
                </span>
              )}
              {incident.resolved_at && (
                <span>
                  {t("status.incidents.resolved")}:{" "}
                  {new Date(incident.resolved_at).toLocaleString()}
                </span>
              )}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
};

export default IncidentHistory;
