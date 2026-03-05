import { useMemo } from "react";
import { clsx } from "clsx";
import type { IncidentModel, StatusPageModel } from "@/api/types.gen";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Trash, CheckCircle, Calendar } from "lucide-react";
import { useLocalizedTranslation } from "@/hooks/useTranslation";

const IncidentCard = ({
  incident,
  statusPage,
  onClick,
  onDelete,
  onResolve,
  isPending,
}: {
  incident: IncidentModel;
  statusPage?: StatusPageModel;
  onClick: () => void;
  onDelete: () => void;
  onResolve: () => void;
  isPending: boolean;
}) => {
  const { t } = useLocalizedTranslation();

  const handleDeleteClick = (e: React.MouseEvent) => {
    e.stopPropagation();
    onDelete();
  };

  const handleResolveClick = (e: React.MouseEvent) => {
    e.stopPropagation();
    onResolve();
  };

  const formatDate = (dateString?: string) => {
    if (!dateString) return null;
    return new Date(dateString).toLocaleString();
  };

  const styleBadgeVariant = useMemo(() => {
    switch (incident.style) {
      case "danger":
        return "destructive" as const;
      case "warning":
        return "default" as const;
      case "info":
        return "secondary" as const;
      default:
        return "outline" as const;
    }
  }, [incident.style]);

  const styleLabel = useMemo(() => {
    switch (incident.style) {
      case "info":
        return t("incidents.style.info");
      case "warning":
        return t("incidents.style.warning");
      case "danger":
        return t("incidents.style.danger");
      default:
        return incident.style;
    }
  }, [incident.style, t]);

  return (
    <Card
      className="mb-2 p-2 hover:cursor-pointer light:hover:bg-gray-100 dark:hover:bg-zinc-800"
      onClick={onClick}
    >
      <CardContent className="px-2">
        <div className="flex justify-between items-center">
          <div className="flex items-center gap-4">
            <div
              className={clsx("flex flex-col min-w-[200px]", {
                "text-gray-500": !incident.active,
              })}
            >
              <h3 className="font-bold mb-1">{incident.title}</h3>
              <div className="flex items-center gap-2">
                <Badge
                  variant={incident.active ? "default" : "outline"}
                  className={clsx({ "text-gray-500": !incident.active })}
                >
                  {incident.active
                    ? t("incidents.active")
                    : t("incidents.resolved")}
                </Badge>
                <Badge variant={styleBadgeVariant}>{styleLabel}</Badge>
                {statusPage?.title && (
                  <Badge variant="outline">{statusPage.title}</Badge>
                )}
              </div>
              <div className="flex items-center gap-4 text-xs text-muted-foreground mt-1">
                {incident.created_at && (
                  <div className="flex items-center gap-1">
                    <Calendar className="h-3 w-3" />
                    <span>{formatDate(incident.created_at)}</span>
                  </div>
                )}
                {incident.resolved_at && (
                  <div className="flex items-center gap-1">
                    <CheckCircle className="h-3 w-3" />
                    <span>
                      {t("incidents.resolved")}:{" "}
                      {formatDate(incident.resolved_at)}
                    </span>
                  </div>
                )}
              </div>
            </div>
          </div>

          <div className="flex items-center gap-2">
            {incident.active && (
              <Button
                variant="ghost"
                size="icon"
                onClick={handleResolveClick}
                className="text-green-500 hover:text-green-700 hover:bg-green-50 dark:hover:bg-green-950"
                aria-label={t("incidents.resolve")}
                disabled={isPending}
              >
                <CheckCircle className="h-4 w-4" />
              </Button>
            )}
            <Button
              variant="ghost"
              size="icon"
              onClick={handleDeleteClick}
              className="text-red-500 hover:text-red-700 hover:bg-red-50 dark:hover:bg-red-950"
              aria-label={t("incidents.delete")}
            >
              <Trash className="h-4 w-4" />
            </Button>
          </div>
        </div>
      </CardContent>
    </Card>
  );
};

export default IncidentCard;
